package registrypkg

import (
	"context"
	"fmt"
	"github.com/eden/go-kratos-pkg/registry/config"
	"github.com/eden/go-kratos-pkg/registry/k8s"
	"github.com/eden/go-kratos-pkg/registry/local"
	"github.com/eden/go-kratos-pkg/registry/nacos"
	"github.com/eden/go-kratos-pkg/registry/util"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
)

type Registry interface {
	registry.Registrar
	registry.Discovery
	Name() string
}

var (
	defaultRegistry Registry
	localRegistry   Registry
)

func Init(conf *config.Registry, _ log.Logger) error {
	r, err := NewLocal(conf)
	if err != nil {
		return err
	}
	localRegistry = r

	if conf != nil && len(conf.LocalServer) > 0 {
		defaultRegistry = localRegistry
		return nil
	}
	if conf != nil && conf.Type == nacos.NACOS {
		r, err := NewNacos(conf)
		if err != nil {
			return err
		}
		defaultRegistry = r
		return nil
	} else {
		r, err := NewK8S(conf)
		if err != nil {
			return err
		}
		defaultRegistry = r
		return nil
	}
}

func NewLocal(conf *config.Registry) (*local.Registry, error) {
	servers := make([]*config.LocalServer, 0)
	if conf != nil {
		servers = append(servers, conf.LocalServer...)
		servers = append(servers, conf.RemoteServer...)
	}
	r := local.NewRegistry(servers)
	return r, nil
}

func NewNacos(conf *config.Registry) (Registry, error) {
	r, err := nacos.NewRegistry(conf, localRegistry)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewK8S(conf *config.Registry) (Registry, error) {
	r, err := k8s.NewRegistry(conf, localRegistry)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func GetRegistrar() registry.Registrar {
	return defaultRegistry
}

func GetDiscovery() registry.Discovery {
	return defaultRegistry
}

func GetServiceDiscoveryEndpoint(serviceName string) string {
	var endpoint string
	if defaultRegistry.Name() == nacos.NACOS {
		endpoint = fmt.Sprintf("discovery:///%s", util.GetGrpcServiceName(serviceName))
	} else {
		endpoint = fmt.Sprintf("discovery:///%s", serviceName)
	}
	return endpoint
}

func IsFixedEndpoint(serviceName string) bool {
	services, _ := localRegistry.GetService(context.Background(), serviceName)
	if len(services) > 0 {
		return true
	}
	return false
}

func GetServiceFixedEndpoint(serviceName string) (string, error) {
	services, err := localRegistry.GetService(context.Background(), serviceName)
	if err != nil {
		return "", err
	}
	if len(services) > 0 {
		service := services[0]
		if len(service.Endpoints) > 0 {
			endpoint := service.Endpoints[0]
			return endpoint, nil
		}
	}
	return "", fmt.Errorf("not found enpdoint: %s", serviceName)
}

// func getServiceEndpoint(serviceName string) (string, error) {
// 	services, err := defaultRegistry.GetService(context.Background(), serviceName)
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(services) > 0 {
// 		service := services[0]
// 		if len(service.Endpoints) > 0 {
// 			endpoint := service.Endpoints[0]
// 			return endpoint, nil
// 		}
// 	}
// 	return "", fmt.Errorf("not found enpdoint: %s", serviceName)
// }
//
// func GetServiceHttpEndpoint(serviceName string) (string, error) {
// 	return getServiceEndpoint(GetHttpServiceName(serviceName))
// }
