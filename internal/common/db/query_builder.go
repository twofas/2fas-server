package db

import (
	"database/sql"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
)

func NewQueryBuilder(database *sql.DB) *goqu.Database {
	return goqu.New("mysql", database)
}
