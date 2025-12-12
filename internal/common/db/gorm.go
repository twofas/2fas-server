package db

import (
	"fmt"

	"gorm.io/gorm/logger"

	"github.com/twofas/2fas-server/config"

	gosql "github.com/go-sql-driver/mysql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGormConnection(conf config.Configuration) *gorm.DB {
	cfg := &gosql.Config{
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%v:%v", conf.Db.Host, conf.Db.Port),
		DBName:               conf.Db.Database,
		User:                 conf.Db.Username,
		Passwd:               conf.Db.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	var logLevel logger.LogLevel

	if conf.Debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	conn, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		panic(err)
	}

	return conn
}
