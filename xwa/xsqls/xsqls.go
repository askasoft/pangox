package xsqls

import (
	"database/sql"
	"errors"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/sqlxlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx/sqlx"
)

var (
	// SDB database instance
	SDB *sqlx.DB

	// DBS database settings
	DBS = map[string]string{}

	// GetErrLogLevels GetErrLogLevel function map
	GetErrLogLevels = map[string]func(error) log.Level{}
)

func Driver() string {
	return DBS["driver"]
}

func Source() string {
	return DBS["source"]
}

func RegisterGetErrLogLevel(driver string, f func(error) log.Level) {
	GetErrLogLevels[driver] = f
}

func OpenDatabase() error {
	sec := ini.GetSection("database")
	if sec == nil {
		return errors.New("missing [database] settings")
	}

	dbs := sec.StringMap()
	if mag.Equal(DBS, dbs) {
		return nil
	}

	driver := sec.GetString("driver")
	source := sec.GetString("source")
	log.Infof("Connect database (%s): %s", driver, source)

	db, err := sql.Open(driver, source)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetConnMaxIdleTime(sec.GetDuration("connMaxIdleTime", 5*time.Minute))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", 10*time.Minute))

	slg := sqlxlog.NewSqlxLogger(
		log.GetLogger("SQL"),
		sec.GetDuration("slowSQL", 2*time.Second),
	)
	slg.GetErrLogLevel = GetErrLogLevels[driver]

	DBS = dbs
	SDB = sqlx.NewDB(db, driver, slg.Trace)

	return nil
}
