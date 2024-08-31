package config

// LocalServer ...
type LocalServer struct {
	Name string
	Addr string
}

// Registry 服务注册发现
type Registry struct {
	LocalServer  []*LocalServer
	RemoteServer []*LocalServer
	Host         string
	Port         int32
	NamespaceId  string
	Username     string
	Password     string
	Type         string
}
