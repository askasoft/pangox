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

type SchemaChange struct {
	Script    string    `gorm:"size:255;not null;primaryKey" json:"script"`
	AppliedAt time.Time `gorm:"not null" json:"applied_at"`
}

func ApplySchemaChanges(db *sqlx.DB, schema string, fsys fs.FS, dir string, loggers ...log.Logger) error {
	logger := getLogger(loggers...)

	des, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
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

	if len(scripts) == 0 {
		return nil
	}

	sort.Strings(scripts)

	applied, err := findAppliedScripts(db, schema)
	if err != nil {
		return err
	}

	for _, script := range scripts {
		if applied.Contains(script) {
			continue
		}

		if err = applyScript(db, schema, fsys, dir, script, logger); err != nil {
			return fmt.Errorf("%s error: %w", script, err)
		}

		dots := ""
		if len(script) < 56 {
			dots = str.Repeat(".", 56-len(script))
		}
		logger.Infof("%s %s OK", script, dots)
	}

	return nil
}

func applyScript(db *sqlx.DB, schema string, fsys fs.FS, dir, script string, logger log.Logger) error {
	sqls, err := fsu.ReadStringFS(fsys, path.Join(dir, script))
	if err != nil {
		return err
	}

	sqls = strings.ReplaceAll(sqls, "SCHEMA", schema)

	err = db.Transaction(func(tx *sqlx.Tx) error {
		sqr := sqx.NewSqlReader(str.NewReader(sqls))

		for i := 1; ; i++ {
			sqs, err := sqr.ReadSql()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			r, err := tx.Exec(sqs)
			if err != nil {
				return err
			}

			cnt, _ := r.RowsAffected()
			logger.Debugf("#%d [%d] = %s", i, cnt, sqs)
		}

		sqs := tx.Rebind(fmt.Sprintf("INSERT INTO %s.schema_changes (script, applied_at) VALUES (?, ?)", schema))
		_, err := tx.Exec(sqs, script, time.Now())
		return err
	})
	return err
}

func findAppliedScripts(tx sqlx.Sqlx, schema string) (*treeset.TreeSet[string], error) {
	rows, err := tx.Query("SELECT script FROM " + schema + ".schema_changes")
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
