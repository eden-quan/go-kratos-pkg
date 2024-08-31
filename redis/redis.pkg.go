package redispkg

import (
	"context"
	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Config struct {
	Addr         string               `json:"addr"`
	Addrs        []string             `json:"addrs"`
	Username     string               `json:"username"`
	DB           int                  `json:"db"`
	Password     string               `json:"password"`
	MaxRetries   int                  `json:"max_retries"`
	DialTimeout  *durationpb.Duration `json:"dial_timeout"`
	ReadTimeout  *durationpb.Duration `json:"read_timeout"`
	WriteTimeout *durationpb.Duration `json:"write_timeout"`
	PoolSize     int                  `json:"pool_size"`
	MinIdleConns int                  `json:"min_idle_conns"`
}

func NewRedisClient(conf *Config, logger log.Logger) redis.UniversalClient {
	lh := log.NewHelper(log.With(logger, "module", "redis"))

	var addrs []string
	if len(conf.Addrs) > 0 {
		addrs = conf.Addrs
	} else {
		addrs = []string{conf.Addr}
	}
	opt := &redis.UniversalOptions{
		Addrs:        addrs,
		DB:           conf.DB,
		Username:     conf.Username,
		Password:     conf.Password,
		MaxRetries:   conf.MaxRetries,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
	}
	if conf.DialTimeout != nil {
		opt.DialTimeout = conf.DialTimeout.AsDuration()
	}
	if conf.ReadTimeout != nil {
		opt.ReadTimeout = conf.ReadTimeout.AsDuration()
	}
	if conf.WriteTimeout != nil {
		opt.WriteTimeout = conf.WriteTimeout.AsDuration()
	}
	client := redis.NewUniversalClient(opt)
	err := client.Ping(context.Background()).Err()
	if err != nil {
		lh.Fatalw("msg", "redis ping failed", "err", err)
	}
	lh.Info("redis successfully connected and ping")
	return client
}
