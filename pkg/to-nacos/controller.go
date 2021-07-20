package tonacos

import (
	"reflect"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
	"github.com/nacos-group/nacos-k8s-sync/pkg/model"
)

type Controller struct {
	nacosClient model.NacosClient

	watchedNamespace string

	serviceInformer cache.SharedIndexInformer
	serviceLister   lister.ServiceLister

	endpointsInformer cache.SharedIndexInformer
	endpointsLister   lister.EndpointsLister

	queue workqueue.RateLimitingInterface

	once sync.Once
}

func NewController(options model.NacosOptions, watchedNamespace string, kubeClient model.KubeClient) (model.Controller, error) {
	nacosClient, err := model.NewNacosClient(options)

	if err != nil {
		return nil, err
	}

	c := &Controller{
		nacosClient:      nacosClient,
		watchedNamespace: watchedNamespace,
	}

	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// list and watch service
	c.serviceInformer = kubeClient.InformerFactory().Core().V1().Services().Informer()
	c.serviceLister = kubeClient.InformerFactory().Core().V1().Services().Lister()
	registerHandlersForInformer(c.serviceInformer, c.queue, c.onServiceEvent)
	// list and watch endpoints
	c.endpointsInformer = kubeClient.InformerFactory().Core().V1().Endpoints().Informer()
	c.endpointsLister = kubeClient.InformerFactory().Core().V1().Endpoints().Lister()
	registerHandlersForInformer(c.endpointsInformer, c.queue, c.onEndpointsEvent)

	return c, nil
}

func (c *Controller) buildAddresses(service *v1.Service, serviceInfo model.ServiceInfo) ([]model.Address, error) {
	endpoints, err := c.endpointsLister.Endpoints(c.watchedNamespace).Get(service.Name)
	if err != nil {
		return nil, err
	}
	addresses := model.ConvertToAddresses(serviceInfo.Port, endpoints)
	return addresses, nil
}

func (c *Controller) onServiceEvent(old, curr interface{}, event model.Event) error {
	currService, ok := curr.(*v1.Service)
	if !ok {
		return nil
	}

	currShouldSync := model.ShouldServiceSync(currService)
	if !currShouldSync && event != model.EventUpdate {
		logger.Infof("Curr Service (%s:%s) should not be synced.", currService.Name, currService.Namespace)
		return nil
	}

	currServiceInfo, err := model.GenerateServiceInfo(currService)
	if err != nil {
		logger.Errorf("Generate curr service info from service (%s:%s) fail.", currService.Name, currService.Namespace)
		return nil
	}

	switch event {
	case model.EventAdd:
		addresses, err := c.buildAddresses(currService, currServiceInfo)
		if err != nil {
			logger.Errorf("Build addresses for curr service (%s:%s) fail, err %v", currServiceInfo.ServiceName, currServiceInfo.Group, err)
			return err
		}
		c.nacosClient.RegisterService(currServiceInfo, addresses)
	case model.EventDelete:
		c.nacosClient.UnregisterService(currServiceInfo)
	case model.EventUpdate:
		oldService, ok := old.(*v1.Service)
		if !ok {
			return nil
		}

		oldServiceInfo, err := model.GenerateServiceInfo(oldService)
		if err != nil {
			logger.Errorf("Generate old service info from service (%s:%s) fail.", oldService.Name, oldService.Namespace)
			return nil
		}

		// Old service should be synced, but now it changed to be not synced.
		// We should Unregister old service.
		if model.ShouldServiceSync(oldService) && !currShouldSync {
			logger.Infof("Old service (%s:%s) should be unregistered.", oldServiceInfo.ServiceName, oldServiceInfo.Group)
			c.nacosClient.UnregisterService(oldServiceInfo)
			return nil
		}

		addresses, err := c.buildAddresses(currService, currServiceInfo)
		if err != nil {
			logger.Errorf("Build addresses for curr service (%s:%s) fail, err %v", currServiceInfo.ServiceName, currServiceInfo.Group, err)
			return err
		}

		// If the service key of old is not equal to new, it means that we get a new service and
		// should to unregister old.
		if oldServiceInfo.ServiceKey != currServiceInfo.ServiceKey {
			// Register new service
			c.nacosClient.RegisterService(currServiceInfo, addresses)
			// Unregister old service
			c.nacosClient.UnregisterService(oldServiceInfo)

		} else if oldServiceInfo.Port != currServiceInfo.Port {
			// If the port of old service is not equal to new, it means that we should push new addresses
			// to nacos and remove old addresses which has old port.
			c.nacosClient.RegisterService(currServiceInfo, addresses)
		} else if !reflect.DeepEqual(oldServiceInfo.Metadata, currServiceInfo.Metadata) {
			// If the metadata of old service is not equal to new, it means that we should republish new
			// address.
			c.nacosClient.RegisterServiceInstances(currServiceInfo, addresses)
		}
	}

	return nil
}

func (c *Controller) onEndpointsEvent(old, curr interface{}, event model.Event) error {
	// The relevant event will be solved by onServiceEvent
	if event == model.EventDelete {
		return nil
	}

	endpoints, ok := curr.(*v1.Endpoints)
	if !ok {
		return nil
	}

	service, err := c.serviceLister.Services(c.watchedNamespace).Get(endpoints.Name)
	if err != nil || service == nil {
		logger.Errorf("Get service (%s:%s) fail.", endpoints.Name, endpoints.Namespace)
		return err
	}

	if !model.ShouldServiceSync(service) {
		logger.Infof("Service (%s:%s) should not be synced.", service.Name, service.Namespace)
		return nil
	}

	serviceInfo, err := model.GenerateServiceInfo(service)
	if err != nil {
		logger.Errorf("Generate service info from service (%s:%s) fail.", service.Name, service.Namespace)
		return nil
	}

	// Publish the newest addresses to nacos.
	addresses, err := c.buildAddresses(service, serviceInfo)
	if err != nil {
		logger.Errorf("Build addresses for service (%s:%s) fail, err %v", serviceInfo.ServiceName, serviceInfo.Group, err)
		return err
	}
	c.nacosClient.RegisterService(serviceInfo, addresses)

	return nil
}

func registerHandlersForInformer(informer cache.SharedIndexInformer, queue workqueue.RateLimitingInterface,
	handler func(interface{}, interface{}, model.Event) error) {

	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				queue.Add(&model.Task{
					Handler: func() error {
						return handler(nil, obj, model.EventAdd)
					},
				})
			},
			UpdateFunc: func(old, cur interface{}) {
				queue.Add(&model.Task{
					Handler: func() error {
						return handler(old, cur, model.EventUpdate)
					},
				})
			},
			DeleteFunc: func(obj interface{}) {
				queue.Add(&model.Task{
					Handler: func() error {
						return handler(nil, obj, model.EventDelete)
					},
				})
			},
		})
}

func (c *Controller) syncAllServiceToNacos() error {
	var err *multierror.Error

	services := c.serviceInformer.GetStore().List()
	for _, service := range services {
		err = multierror.Append(err, c.onServiceEvent(nil, service, model.EventAdd))
	}

	return multierror.Flatten(err.ErrorOrNil())
}

func (c *Controller) HasSynced() bool {
	if !c.serviceInformer.HasSynced() || !c.endpointsInformer.HasSynced() {
		return false
	}

	t0 := time.Now()
	c.once.Do(func() {
		if err := c.syncAllServiceToNacos(); err != nil {
			return
		}
	})

	logger.Infof("Have Synced all services to nacos, cost %s.", time.Since(t0))
	return true
}

func (c *Controller) processQueueTask() {
	obj, shutdown := c.queue.Get()
	defer c.queue.Done(obj)

	if shutdown {
		return
	}

	task, ok := obj.(*model.Task)
	if !ok {
		logger.Warn("Convert to task fail.")
		return
	}

	if err := task.Handler(); err != nil {
		if c.queue.NumRequeues(obj) < model.MaxRetry {
			time.AfterFunc(model.DefaultTaskDelay, func() {
				logger.Warnf("Task handle fail and put into queue again, err %v", err)
				c.queue.AddRateLimited(obj)
			})
		} else {
			logger.Warn("Task handle retry reach max times.")
			c.queue.Forget(obj)
		}
	}
}

func (c *Controller) Run(stop <-chan struct{}) {
	defer c.queue.ShutDown()

	cache.WaitForCacheSync(stop, c.HasSynced)

	wait.Until(c.processQueueTask, 0, stop)
}
