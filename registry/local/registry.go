package local

import (
	"context"
	"github.com/eden-quan/go-kratos-pkg/registry/config"
	"github.com/eden-quan/go-kratos-pkg/registry/util"
	"github.com/go-kratos/kratos/v2/registry"
)

var _ registry.Discovery = (*Registry)(nil)
var _ registry.Registrar = (*Registry)(nil)

const LOCAL = "local"

type Registry struct {
	services map[string]map[string]*registry.ServiceInstance
}

func NewRegistry(servers []*config.LocalServer) *Registry {
	r := &Registry{
		services: make(map[string]map[string]*registry.ServiceInstance),
	}
	for _, s := range servers {
		r.addServiceInstance(r.newServiceInstance(s.Name, s.Addr))
		r.addServiceInstance(r.newServiceInstance(util.GetGrpcServiceName(s.Name), s.Addr))
	}
	return r
}

func (r *Registry) Name() string {
	return LOCAL
}

func (r *Registry) newServiceInstance(name string, addr string) *registry.ServiceInstance {
	service := &registry.ServiceInstance{
		ID:        name,
		Name:      name,
		Version:   name,
		Metadata:  make(map[string]string),
		Endpoints: []string{addr},
	}
	return service
}

func (r *Registry) addServiceInstance(service *registry.ServiceInstance) {
	if _, ok := r.services[service.Name]; !ok {
		r.services[service.Name] = make(map[string]*registry.ServiceInstance)
	}
	r.services[service.Name][service.ID] = service
}

func (r *Registry) deleteServiceInstance(service *registry.ServiceInstance) {
	if _, ok := r.services[service.Name]; ok {
		delete(r.services[service.Name], service.ID)
		if len(r.services[service.Name]) == 0 {
			delete(r.services, service.Name)
		}
	}
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	r.addServiceInstance(service)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	r.deleteServiceInstance(service)
	return nil
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	if services, ok := r.services[serviceName]; ok {
		list := make([]*registry.ServiceInstance, 0, len(services))
		for _, service := range services {
			list = append(list, service)
		}
		return list, nil
	}
	return nil, nil
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newWatcher(r, serviceName), nil
}

func (r *Registry) Exist(serviceName string) bool {
	if _, ok := r.services[serviceName]; ok {
		return true
	}
	return false
}

type watcher struct {
	registry    *Registry
	serviceName string
	blockCh     chan struct{}
	stopped     bool
}

func newWatcher(registry *Registry, serviceName string) *watcher {
	w := &watcher{
		registry:    registry,
		serviceName: serviceName,
		blockCh:     make(chan struct{}, 1),
	}
	w.blockCh <- struct{}{} // 第一次Next返回
	return w
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	if w.stopped {
		return nil, context.Canceled
	}
	_, ok := <-w.blockCh
	if ok {
		return w.registry.GetService(context.Background(), w.serviceName)
	}
	return nil, context.Canceled
}

func (w *watcher) Stop() error {
	w.stopped = true
	close(w.blockCh)
	return nil
}
