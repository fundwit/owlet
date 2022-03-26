package testinfra

import (
	"database/sql/driver"
	"owlet/server/infra/persistence"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fundwit/go-commons/types"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SetUpMockSql() (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mysqlConfig := mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}
	gormDB, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	persistence.ActiveGormDB = gormDB
	return gormDB, mock
}

type AnyArgument struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyArgument) Match(v driver.Value) bool {
	return true
}

type AnyPastTime struct {
	Range time.Duration
}

// Match satisfies sqlmock.Argument interface
func (a AnyPastTime) Match(v driver.Value) bool {
	str, ok := v.(string)
	if !ok {
		return false
	}
	t := types.Timestamp{}
	if err := t.Scan(str); err != nil {
		return false
	}
	return time.Since(t.Time()) < a.Range
}

type AnyId struct{}

func (a AnyId) Match(v driver.Value) bool {
	id, ok := v.(int64)
	if !ok {
		return false
	}
	return id > 0
}
