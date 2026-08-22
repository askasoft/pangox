package xsqls

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/askasoft/pango/cog/treeset"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

type Migration struct {
	Script    string    `gorm:"size:255;not null;primaryKey" json:"script"`
	AppliedAt time.Time `gorm:"not null" json:"applied_at"`
}

type Migrator struct {
	sdb    *sqlx.DB
	schema string
	table  string
}

func NewMigrator(sdb *sqlx.DB, schema, table string) *Migrator {
	return &Migrator{sdb: sdb, schema: schema, table: table}
}

func InitMigrations(sdb *sqlx.DB, schema, table string, fsys fs.FS, dir string) error {
	return NewMigrator(sdb, schema, table).InitMigrations(fsys, dir)
}

func ApplyMigrations(sdb *sqlx.DB, schema, table string, fsys fs.FS, dir string, loggers ...log.Logger) error {
	return NewMigrator(sdb, schema, table).ApplyMigrations(fsys, dir)
}

func (m *Migrator) CreateMigration(tx sqlx.Sqlx, script string, appliedAt time.Time) error {
	sqs := tx.Rebind("INSERT INTO " + m.tableName() + " (script, applied_at) VALUES (?, ?)")
	_, err := tx.Update(sqs, script, appliedAt)
	return err
}

func (m *Migrator) InitMigrations(fsys fs.FS, dir string) error {
	scripts, err := m.findMigrations(fsys, dir)
	if err != nil {
		return err
	}

	err = m.sdb.Transaction(func(tx *sqlx.Tx) error {
		for _, s := range scripts {
			if err := m.CreateMigration(tx, s, time.Now()); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (m *Migrator) ApplyMigrations(fsys fs.FS, dir string, loggers ...log.Logger) error {
	logger := getLogger(loggers...)

	scripts, err := m.findMigrations(fsys, dir)
	if err != nil {
		return err
	}

	applied, err := m.findAppliedScripts(m.sdb)
	if err != nil {
		return err
	}

	for _, script := range scripts {
		if applied.Contains(script) {
			continue
		}

		if err = m.applyScript(fsys, dir, script); err != nil {
			return fmt.Errorf("%s error: %w", script, err)
		}

		dots := ""
		if len(script) < 56 {
			dots = str.Repeat(".", 56-len(script))
		}
		logger.Warnf("%s %s OK", script, dots)
	}

	return nil
}

func (m *Migrator) tableName() string {
	if m.schema == "" {
		return m.table
	}
	return m.schema + "." + m.table
}

func (m *Migrator) findMigrations(fsys fs.FS, dir string) ([]string, error) {
	des, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	var scripts []string
	for _, de := range des {
		if de.IsDir() {
			continue
		}

		ext := path.Ext(de.Name())
		if strings.HasSuffix(ext, ".sql") {
			scripts = append(scripts, path.Base(de.Name()))
		}
	}

	sort.Strings(scripts)
	return scripts, nil
}

func (m *Migrator) findAppliedScripts(tx sqlx.Sqlx) (*treeset.TreeSet[string], error) {
	rows, err := tx.Query("SELECT script FROM " + m.tableName())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scripts := treeset.NewTreeSet(str.CompareFold)

	var script string
	for rows.Next() {
		if err = rows.Scan(&script); err != nil {
			return nil, err
		}
		scripts.Add(script)
	}

	return scripts, nil
}

func (m *Migrator) applyScript(fsys fs.FS, dir, script string) error {
	sqls, err := fsu.ReadStringFS(fsys, path.Join(dir, script))
	if err != nil {
		return err
	}

	sqls = strings.ReplaceAll(sqls, "SCHEMA", m.schema)

	err = m.sdb.Transaction(func(tx *sqlx.Tx) error {
		sqr := sqx.NewSqlReader(str.NewReader(sqls))

		for i := 1; ; i++ {
			sqs, err := sqr.ReadSql()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			_, err = tx.Update(sqs)
			if err != nil {
				return err
			}
		}

		return m.CreateMigration(tx, script, time.Now())
	})
	return err
}
