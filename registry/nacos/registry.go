package nacos

import (
	"context"
	"github.com/eden-quan/go-kratos-pkg/registry/config"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ registry.Discovery = (*Registry)(nil)
var _ registry.Registrar = (*Registry)(nil)

const NACOS = "nacos"

type Registry struct {
	registry *nacos.Registry
	// local    registry.Discovery
}

func NewRegistry(conf *config.Registry, local registry.Discovery) (*Registry, error) {
	// servers := make([]*config.Registry_LocalServer, 0)
	if conf != nil {
		if conf.Host == "" {
			conf.Host = "127.0.0.1"
		}
		if conf.Port == 0 {
			conf.Port = 8848
		}
		// servers = append(servers, conf.LocalServer...)
		// servers = append(servers, conf.RemoteServer...)
	}
	// localDiscovery := local.NewRegistry(servers)

	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ServerConfigs: []constant.ServerConfig{
				{
					Scheme:      constant.DEFAULT_SERVER_SCHEME,
					ContextPath: constant.DEFAULT_CONTEXT_PATH,
					IpAddr:      conf.Host,
					Port:        uint64(conf.Port),
				},
			},
			ClientConfig: &constant.ClientConfig{
				NamespaceId:          conf.NamespaceId,
				TimeoutMs:            5000,
				BeatInterval:         3000,
				LogDir:               "logs",
				NotLoadCacheAtStart:  true,
				UpdateCacheWhenEmpty: true,
				LogLevel:             "info",
				Username:             conf.Username,
				Password:             conf.Password,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	nc := nacos.New(client)

	r := &Registry{
		registry: nc,
		// local:    local,
	}
	return r, nil
}

func (r *Registry) Name() string {
	return NACOS
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
	return r.registry.Watch(ctx, serviceName)
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	// if services, _ := r.local.GetService(ctx, serviceName); len(services) > 0 {
	// 	return services, nil
	// }
	return r.registry.GetService(ctx, serviceName)
}
