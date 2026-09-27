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
	"github.com/askasoft/pango/str"
)

type dbsrc struct {
	sdb *sqlx.DB
	dbc map[string]string
	slg *sqlxlog.SqlxLogger
}

const (
	AnyDriver = "*"
)

var (
	// GetErrLogLevels GetErrLogLevel function map
	GetErrLogLevels = map[string]func(error) log.Level{}

	// GetSQLLogLevels GetSQLLogLevel function map
	GetSQLLogLevels = map[string]func(string) log.Level{}

	// IsSlowSQLs IsSlowSQL function map
	IsSlowSQLs = map[string]func(string, time.Duration) bool{}

	// database sources
	sources = map[string]*dbsrc{}
)

func init() {
	RegisterIsSlowSQL(AnyDriver, defaultSlowSQLLog)
}

func RegisterGetErrLogLevel(driver string, f func(error) log.Level) {
	GetErrLogLevels[driver] = f
}

func RegisterGetSQLLogLevel(driver string, f func(string) log.Level) {
	GetSQLLogLevels[driver] = f
}

func RegisterIsSlowSQL(driver string, f func(string, time.Duration) bool) {
	IsSlowSQLs[driver] = f
}

func getRegisteredFunction[T any](m map[string]T, k string) T {
	if v, ok := m[k]; ok {
		return v
	}
	return m[AnyDriver]
}

func defaultSlowSQLLog(sql string, _ time.Duration) bool {
	sql = str.StripLeft(sql)
	return !str.StartsWithFold(sql, "ALTER") && !str.StartsWithFold(sql, "VACUUM")
}

func SDB(id ...string) *sqlx.DB {
	if src, ok := sources[asg.First(id)]; ok {
		return src.sdb
	}
	return nil
}

func Driver(id ...string) string {
	return config("driver", id...)
}

func Source(id ...string) string {
	return config("source", id...)
}

func Logger(id ...string) *sqlxlog.SqlxLogger {
	if src, ok := sources[asg.First(id)]; ok {
		return src.slg
	}
	return nil
}

func config(key string, id ...string) string {
	if src, ok := sources[asg.First(id)]; ok {
		if val, ok := src.dbc[key]; ok {
			return val
		}
	}
	return ""
}

func OpenDatabase(id ...string) error {
	return openDatabase(asg.First(id))
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
	if src, ok := sources[id]; ok {
		if mag.Equal(dbc, src.dbc) {
			return nil
		}
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

	slg := sqlxlog.NewSqlxLogger(log.GetLogger("SQL"))
	slg.SlowSQLTime = sec.GetDuration("slowSQL", 2*time.Second)
	slg.GetErrLogLevel = getRegisteredFunction(GetErrLogLevels, driver)
	slg.GetSQLLogLevel = getRegisteredFunction(GetSQLLogLevels, driver)
	slg.IsSlowSQL = getRegisteredFunction(IsSlowSQLs, driver)

	sources[id] = &dbsrc{
		sdb: sqlx.NewDB(db, driver, slg.Trace),
		dbc: dbc,
		slg: slg,
	}
	return nil
}
