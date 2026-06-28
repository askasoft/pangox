package xsqls

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/sqlxlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx/sqlx"
)

type dbcfg = map[string]string

var (
	// GetErrLogLevels GetErrLogLevel function map
	GetErrLogLevels = map[string]func(error) log.Level{}

	// sdbs database instances
	sdbs = map[string]*sqlx.DB{}

	// dbcs database configurations
	dbcs = map[string]dbcfg{}
)

func RegisterGetErrLogLevel(driver string, f func(error) log.Level) {
	GetErrLogLevels[driver] = f
}

func SDB(ids ...string) *sqlx.DB {
	id := asg.First(ids)
	return sdbs[id]
}

func Driver(ids ...string) string {
	return config("driver", ids...)
}

func Source(name ...string) string {
	return config("source", name...)
}

func config(key string, ids ...string) string {
	id := asg.First(ids)
	if dbc, ok := dbcs[id]; ok {
		return dbc[key]
	}
	return ""
}

func OpenDatabase(ids ...string) error {
	id := asg.First(ids)
	return openDatabase(id)
}

func OpenDatabases(ids ...string) error {
	for _, id := range ids {
		if err := openDatabase(id); err != nil {
			return err
		}
	}
	return nil
}

func openDatabase(id string) error {
	key := "database"
	if id != "" {
		key += "." + id
	}

	sec := ini.GetSection(key)
	if sec == nil {
		return fmt.Errorf("missing [%s] settings", key)
	}

	dbc := sec.StringMap()
	if mag.Equal(dbc, dbcs[id]) {
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

	dbcs[id] = dbc
	sdbs[id] = sqlx.NewDB(db, driver, slg.Trace)
	return nil
}
