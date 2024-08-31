package mysqlpkg

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Model 是数据库表的基本字段
type Model struct {
	ID         int64     `db:"id"`          // 主键
	IsDeleted  bool      `db:"is_deleted"`  // 是否删除
	CreateTime time.Time `db:"create_time"` // 创建时间
	UpdateTime time.Time `db:"update_time"` // 更新时间
}

type Config struct {
	Addr            string               `json:"addr"`
	DbName          string               `json:"db_name"`
	Debug           bool                 `json:"debug"`
	ConnMaxLifetime *durationpb.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime *durationpb.Duration `json:"conn_max_idle_time"`
	MaxIdleConns    int                  `json:"max_idle_conns"`
	MaxOpenConns    int                  `json:"max_open_conns"`
}

func NewMysqlClient(config *Config, logger log.Logger) *sqlx.DB {
	logHelper := log.NewHelper(log.With(logger, "module", "mysql"))

	// custom dsn string
	// dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?parseTime=true",
	// 	config.User,
	// 	config.Password,
	// 	config.Host,
	// 	config.Port,
	// 	config.DbName,
	// )

	db, err := sqlx.Open("mysql", config.Addr)
	if err != nil {
		logHelper.Fatalw("msg", "mysql connect failed", "err", err)
	}
	if err1 := db.Ping(); err1 != nil {
		logHelper.Fatalw("msg", "mysql ping failed", "err", err1)
	}

	if config.ConnMaxIdleTime != nil {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime.AsDuration())
	}
	if config.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(config.ConnMaxLifetime.AsDuration())
	}
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	return db
}
