package mysqlpkg

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func TestNewMysqlClient(t *testing.T) {
	config := Config{
		Addr:   "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
		DbName: "dbname",
	}
	mysqlClient := NewMysqlClient(&config, log.With(log.DefaultLogger, "module", "mysql"))
	t.Log(mysqlClient.Ping())
}
