package redispkg

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func TestRedis(t *testing.T) {
	config := &Config{
		Addr:     "r-wz92wby4bez2bb83fgpd.redis.rds.aliyuncs.com:6379",
		Password: "QSUBLQy5CYWj2wil",
	}

	client := NewRedisClient(config, log.DefaultLogger)
	_ = client
}
