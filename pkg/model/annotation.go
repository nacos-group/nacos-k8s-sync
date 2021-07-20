package model

import (
	"encoding/json"
	"strconv"

	v1 "k8s.io/api/core/v1"

	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
)

const (
	// annotationServiceSync is the key of the annotation that determines
	// whether to sync the Service resource or not.
	annotationServiceSync = "nacos.alibaba.com/service-sync"

	// annotationServiceName is set to override the name of the service
	// registered.
	annotationServiceName = "nacos.alibaba.com/service-name"

	// annotationServiceGroup is set to override the group of the service
	// registered.
	annotationServiceGroup = "nacos.alibaba.com/service-group"

	// annotationServicePort specifies the port to use as the service instance
	// port when registering a service. This can be a named port in the
	// service or an integer value.
	annotationServicePort = "nacos.alibaba.com/service-port"

	// annotationServiceMeta specifies the meta of nacos service.
	// The format must be json.
	annotationServiceMeta = "nacos.alibaba.com/service-meta"
)

func ShouldServiceSync(svc *v1.Service) bool {
	raw, ok := svc.Annotations[annotationServiceSync]
	if !ok {
		return false
	}

	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}

	return v
}

func GenerateServiceInfo(svc *v1.Service) (ServiceInfo, error) {
	serviceName := svc.Annotations[annotationServiceName]
	if serviceName == "" {
		// fall back to get the name of service resource
		logger.Info("The service name annotion is empty, so we use the name of service resource.")
		serviceName = svc.Name
	}

	port, err := strconv.ParseUint(svc.Annotations[annotationServicePort], 0, 0)
	if err != nil {
		return ServiceInfo{}, err
	}

	var meta map[string]string
	rawMeta := svc.Annotations[annotationServiceMeta]
	if rawMeta != "" {
		if err := json.Unmarshal([]byte(svc.Annotations[annotationServiceMeta]), &meta); err != nil {
			return ServiceInfo{}, err
		}
	}

	// Now we only trust the annotations.
	// TODO Extract value from the spec of service resource for extended features
	return ServiceInfo{
		ServiceKey: ServiceKey{
			ServiceName: serviceName,
			Group:       svc.Annotations[annotationServiceGroup],
		},
		Port:     port,
		Metadata: meta,
	}, nil
}
