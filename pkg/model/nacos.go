package model

import (
	"os"
	"path"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	v1 "k8s.io/api/core/v1"

	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
)

type NacosOptions struct {
	Namespace string

	// ServersIP are explicitly specified to be connected to nacos by client.
	ServersIP []string

	// ServerPort are explicitly specified to be used when the client connects to nacos.
	ServerPort uint64
}

func ConvertToNacosClientParam(options NacosOptions) vo.NacosClientParam {
	clientConfig := constant.ClientConfig{
		NamespaceId:         options.Namespace,
		NotLoadCacheAtStart: true,
		LogDir:              path.Join(os.Getenv("HOME"), "logs", "nacos-go-sdk"),
	}

	var serversConfig []constant.ServerConfig
	for _, ip := range options.ServersIP {
		serversConfig = append(serversConfig, constant.ServerConfig{
			IpAddr: ip,
			Port:   options.ServerPort,
		})
	}

	return vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serversConfig,
	}
}

type ServiceKey struct {
	ServiceName string

	Group string
}

type ServiceInfo struct {
	ServiceKey

	Port uint64

	Metadata map[string]string
}

type NacosClient interface {
	RegisterService(ServiceInfo, []Address)

	UnregisterService(ServiceInfo)

	RegisterServiceInstances(serviceInfo ServiceInfo, addresses []Address)

	UnregisterServiceInstances(serviceInfo ServiceInfo, addresses []Address)
}

type nacosClient struct {
	client      naming_client.INamingClient
	servicesMap map[ServiceKey][]Address
}

func NewNacosClient(options NacosOptions) (NacosClient, error) {
	nacosConfig := ConvertToNacosClientParam(options)
	client, err := clients.NewNamingClient(nacosConfig)
	if err != nil {
		return nil, err
	}

	return &nacosClient{
		client:      client,
		servicesMap: make(map[ServiceKey][]Address),
	}, nil
}

func (c *nacosClient) RegisterService(serviceInfo ServiceInfo, addresses []Address) {
	old := c.servicesMap[serviceInfo.ServiceKey]
	added, deleted := diffAddresses(old, addresses)
	logger.Infof("Register service (%s@@%s), added %d, deleted %d.",
		serviceInfo.ServiceName, serviceInfo.Group, len(added), len(deleted))

	c.RegisterServiceInstances(serviceInfo, added)
	c.UnregisterServiceInstances(serviceInfo, deleted)

	c.servicesMap[serviceInfo.ServiceKey] = addresses
}

func (c *nacosClient) UnregisterService(serviceInfo ServiceInfo) {
	logger.Infof("Unregister service (%s@@%s).", serviceInfo.ServiceName, serviceInfo.Group)
	c.UnregisterServiceInstances(serviceInfo, c.servicesMap[serviceInfo.ServiceKey])
	delete(c.servicesMap, serviceInfo.ServiceKey)
}

func (c *nacosClient) RegisterServiceInstances(serviceInfo ServiceInfo, addresses []Address) {
	for _, address := range addresses {
		if _, err := c.client.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          address.IP,
			Port:        address.Port,
			Weight:      DefaultNacosEndpointWeight,
			Enable:      true,
			Healthy:     true,
			Metadata:    serviceInfo.Metadata,
			ServiceName: serviceInfo.ServiceName,
			GroupName:   serviceInfo.Group,
			Ephemeral:   true,
		}); err != nil {
			logger.Errorf("Register instance (%s:%d) with service (%s@@%s) fail, err %v.",
				address.IP, address.Port, serviceInfo.ServiceName, serviceInfo.Group, err)
		}
	}
}

func (c *nacosClient) UnregisterServiceInstances(serviceInfo ServiceInfo, addresses []Address) {
	for _, address := range addresses {
		if _, err := c.client.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          address.IP,
			Port:        address.Port,
			ServiceName: serviceInfo.ServiceName,
			GroupName:   serviceInfo.Group,
			Ephemeral:   true,
		}); err != nil {
			logger.Errorf("Unregister instance (%s:%d) with service (%s@@%s) fail, err %v.",
				address.IP, address.Port, serviceInfo.ServiceName, serviceInfo.Group, err)
		}
	}
}

type Address struct {
	IP   string `json:"ip"`
	Port uint64 `json:"port"`
}

func diffAddresses(old, curr []Address) ([]Address, []Address) {
	var added, deleted []Address
	oldAddressesSet := make(map[Address]struct{}, len(old))
	newAddressesSet := make(map[Address]struct{}, len(curr))

	for _, s := range old {
		oldAddressesSet[s] = struct{}{}
	}
	for _, s := range curr {
		newAddressesSet[s] = struct{}{}
	}

	for oldAddress := range oldAddressesSet {
		if _, exist := newAddressesSet[oldAddress]; !exist {
			deleted = append(deleted, oldAddress)
		}
	}

	for newAddress := range newAddressesSet {
		if _, exist := oldAddressesSet[newAddress]; !exist {
			added = append(added, newAddress)
		}
	}

	return added, deleted
}

func ConvertToAddresses(realPort uint64, endpoints *v1.Endpoints) []Address {
	var addresses []Address
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				if port.Port == int32(realPort) {
					addresses = append(addresses, Address{
						IP:   address.IP,
						Port: realPort,
					})
				}
			}
		}
	}

	return addresses
}
