package k8s

import (
	"context"
	"fmt"
	"github.com/eden/go-kratos-pkg/registry/config"
	"github.com/go-kratos/kratos/v2/registry"
)

var _ registry.Discovery = (*Registry)(nil)
var _ registry.Registrar = (*Registry)(nil)

const K8S = "k8s"

type Registry struct {
	discovery registry.Discovery
	registry  registry.Registrar
	// local     registry.Discovery
}

func NewRegistry(conf *config.Registry, local registry.Discovery) (*Registry, error) {
	// servers := make([]*config.Registry_LocalServer, 0)
	// if conf != nil {
	// 	servers = append(servers, conf.LocalServer...)
	// 	servers = append(servers, conf.RemoteServer...)
	// }
	//
	// localDiscovery := local.NewRegistry(servers)
	r := &Registry{
		// local: local,
	}

	if discovery, err := getDiscovery(); err != nil {
		return nil, fmt.Errorf("getDiscovery:%s", err.Error())
	} else {
		r.discovery = discovery
	}
	if registrar, err := getRegistrar(); err != nil {
		return nil, fmt.Errorf("getRegistrar:%s", err.Error())
	} else {
		r.registry = registrar
	}
	return r, nil
}

func (r *Registry) Name() string {
	return K8S
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	return r.registry.Register(ctx, service)
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	return r.registry.Deregister(ctx, service)
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	// if services, _ := r.local.GetService(ctx, serviceName); len(services) > 0 {
	// 	return r.local.Watch(ctx, serviceName)
	// }
	return r.discovery.Watch(ctx, serviceName)
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	// if services, _ := r.local.GetService(ctx, serviceName); len(services) > 0 {
	// 	return services, nil
	// }
	return r.discovery.GetService(ctx, serviceName)
}
