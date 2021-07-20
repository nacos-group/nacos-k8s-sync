package bootstrap

import (
	"fmt"

	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
	"github.com/nacos-group/nacos-k8s-sync/pkg/model"
	tonacos "github.com/nacos-group/nacos-k8s-sync/pkg/to-nacos"
)

type Options struct {
	KubeOptions model.KubeOptions

	NacosOptions model.NacosOptions

	Direction model.Direction
}

type Server struct {
	toNacosController model.Controller

	toK8sController model.Controller

	kubeClient model.KubeClient
}

func NewServer(options Options) (*Server, error) {
	server := &Server{}

	if err := server.initKubeClient(options.KubeOptions); err != nil {
		return nil, err
	}

	if err := server.initController(options); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) initKubeClient(options model.KubeOptions) error {
	kubeClient, err := model.NewKubeClient(options)
	if err != nil {
		logger.Error("Init kube client fail.")
		return err
	}

	s.kubeClient = kubeClient
	return nil
}

func (s *Server) initController(options Options) error {
	switch options.Direction {
	case model.ToNacos:
		tonacosController, err := tonacos.NewController(options.NacosOptions, options.KubeOptions.WatchedNamespace, s.kubeClient)
		if err != nil {
			logger.Error("Init to nacos controller fail.")
			return err
		}
		s.toNacosController = tonacosController
	default:
		return fmt.Errorf("not supported type direction %s", options.Direction)
	}

	return nil
}

func (s *Server) Run(stop <-chan struct{}) {
	go s.kubeClient.Run(stop)

	if s.toNacosController != nil {
		go s.toNacosController.Run(stop)
	}

	if s.toK8sController != nil {
		go s.toK8sController.Run(stop)
	}
}
