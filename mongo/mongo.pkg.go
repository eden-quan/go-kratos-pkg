package mongopkg

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Config struct {
	AppName           string               `json:"app_name"`
	Addr              string               `json:"addr"`
	MaxPoolSize       uint64               `json:"max_pool_size"`
	MinPoolSize       uint64               `json:"min_pool_size"`
	MaxConnecting     uint64               `json:"max_connecting"`
	ConnectTimeout    *durationpb.Duration `json:"connect_timeout"`
	HeartbeatInterval *durationpb.Duration `json:"heartbeat_interval"`
	MaxConnIdleTime   *durationpb.Duration `json:"max_conn_idle_time"`
	Timeout           *durationpb.Duration `json:"timeout"`
	Hosts             []string             `json:"hosts"`
	Debug             bool                 `json:"debug"`
}

func NewMongoClient(config *Config, logger log.Logger) *mongo.Client {
	lh := log.NewHelper(log.With(logger, "module", "mongo"))

	clientOpt := options.Client()
	clientOpt.SetHosts(config.Hosts)
	if config.Addr != "" {
		clientOpt.ApplyURI(config.Addr)
	}
	if config.ConnectTimeout != nil {
		clientOpt.SetConnectTimeout(config.ConnectTimeout.AsDuration())
	}
	if config.HeartbeatInterval != nil {
		clientOpt.SetHeartbeatInterval(config.HeartbeatInterval.AsDuration())
	}
	if config.MaxConnIdleTime != nil {
		clientOpt.SetMaxConnIdleTime(config.MaxConnIdleTime.AsDuration())
	}
	if config.Timeout != nil {
		clientOpt.SetTimeout(config.Timeout.AsDuration())
	}
	clientOpt.SetMaxPoolSize(config.MaxPoolSize)
	clientOpt.SetMinPoolSize(config.MinPoolSize)
	clientOpt.SetMaxConnecting(config.MaxConnecting)
	if config.Debug {
		clientOpt.SetMonitor(NewMonitor(lh))
	}

	client, err := mongo.Connect(context.Background(), clientOpt)
	if err != nil {
		lh.Fatalw("msg", "mongo connect failed", "err", err)
	}
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		lh.Fatalw("msg", "mongo ping failed", "err", err)
	}
	lh.Info("Mongodb successfully connected and ping")
	return client
}
