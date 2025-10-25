package mysqlxsm

import (
	"errors"

	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox/xsm"
	"github.com/askasoft/pangox/xsm/mysm"
)

type ssm struct {
	db *sqlx.DB
}

func SM(db *sqlx.DB) xsm.SchemaManager {
	return &ssm{db}
}

func (ssm *ssm) GetSchema(s string) (*xsm.SchemaInfo, error) {
	if mysm.IsSysDB(s) {
		return nil, nil
	}

	sqb := ssm.db.Builder()
	sqb.Select(
		"schema_name AS name",
		"(SELECT SUM(data_length + index_length) FROM information_schema.tables WHERE table_schema = schema_name) AS size",
		"schema_comment AS comment",
	)
	sqb.From("information_schema.schemata")
	sqb.Eq("schema_name", s)

	sql, args := sqb.Build()

	schema := &xsm.SchemaInfo{}
	if err := ssm.db.Get(schema, sql, args...); err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return schema, nil
}

func (ssm *ssm) ExistsSchema(s string) (bool, error) {
	if mysm.IsSysDB(s) {
		return false, nil
	}

	sqb := ssm.db.Builder()
	sqb.Select("schema_name")
	sqb.From("information_schema.schemata")
	sqb.Eq("schema_name", s)
	sql, args := sqb.Build()

	var sn string
	err := ssm.db.Get(&sn, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ssm *ssm) ListSchemas() ([]string, error) {
	sqb := ssm.db.Builder()
	sqb.Select("schema_name").From("information_schema.schemata")
	sqb.NotIn("schema_name", mysm.SysDBs)
	sqb.Order("schema_name")
	sql, args := sqb.Build()

	rows, err := ssm.db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sn string

	var ss []string
	for rows.Next() {
		if err = rows.Scan(&sn); err != nil {
			return nil, err
		}
		ss = append(ss, sn)
	}
	return ss, nil
}

func (ssm *ssm) CreateSchema(name, comment string) error {
	err := ssm.db.Transaction(func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(mysm.SQLCreateSchema(name)); err != nil {
			return err
		}
		if comment != "" {
			if _, err := tx.Exec(mysm.SQLCommentSchema(name, comment)); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (ssm *ssm) CommentSchema(name string, comment string) error {
	_, err := ssm.db.Exec(mysm.SQLCommentSchema(name, comment))
	return err
}

func (ssm *ssm) RenameSchema(old string, new string) error {
	return errors.New("mysql: rename schema is unsupported")
}

func (ssm *ssm) DeleteSchema(name string) error {
	_, err := ssm.db.Exec(mysm.SQLDeleteSchema(name))
	return err
}

func (ssm *ssm) addQuery(sqb *sqlx.Builder, sq *xsm.SchemaQuery) {
	sqb.NotIn("schema_name", mysm.SysDBs)
	if sq.Name != "" {
		sqb.Like("schema_name", sqx.StringLike(sq.Name))
	}
}

func (ssm *ssm) CountSchemas(sq *xsm.SchemaQuery) (total int, err error) {
	sqb := ssm.db.Builder()
	sqb.Count()
	sqb.From("information_schema.schemata")
	ssm.addQuery(sqb, sq)
	sql, args := sqb.Build()

	err = ssm.db.Get(&total, sql, args...)
	return
}

func (ssm *ssm) FindSchemas(sq *xsm.SchemaQuery) (schemas []*xsm.SchemaInfo, err error) {
	sqb := ssm.db.Builder()
	sqb.Select(
		"schema_name AS name",
		"(SELECT SUM(data_length + index_length) FROM information_schema.tables WHERE table_schema = schema_name) AS size",
		"schema_comment AS comment",
	)
	sqb.From("information_schema.schemata")
	ssm.addQuery(sqb, sq)

	sqb.Orders(sq.Order, "name")
	sqb.Offset(sq.Start()).Limit(sq.Limit)

	sql, args := sqb.Build()

	err = ssm.db.Select(&schemas, sql, args...)
	return
}
