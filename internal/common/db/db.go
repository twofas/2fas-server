package db

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/twofas/2fas-server/config"
)

func NewDbConnection(conf config.Configuration) *sql.DB {
	cfg := &mysql.Config{
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%v:%v", conf.Db.Host, conf.Db.Port),
		DBName:               conf.Db.Database,
		User:                 conf.Db.Username,
		Passwd:               conf.Db.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		MultiStatements:      true,
	}

	conn, err := sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		panic(err)
	}

	return conn
}
