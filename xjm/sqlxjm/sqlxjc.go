package sqlxjm

import (
	"errors"
	"time"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox/xjm"
)

type sjc struct {
	db sqlx.Sqlx
	tb string // job chain table
}

func JC(db sqlx.Sqlx, table string) xjm.JobChainer {
	return &sjc{
		db: db,
		tb: table,
	}
}

func (sjc *sjc) GetJobChain(cid int64) (*xjm.JobChain, error) {
	sqb := sjc.db.Builder()
	sqb.Select().From(sjc.tb).Where("id = ?", cid)
	sql, args := sqb.Build()

	jc := &xjm.JobChain{}
	err := sjc.db.Get(jc, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return nil, xjm.ErrJobChainMissing
		}
		return nil, err
	}
	return jc, nil
}

func (sjc *sjc) findJobChains(name string, start, limit int, asc bool, status ...string) *sqlx.Builder {
	sqb := sjc.db.Builder()

	sqb.Select().From(sjc.tb)
	if name != "" {
		sqb.Where("name = ?", name)
	}
	if len(status) > 0 {
		sqb.In("status", status)
	}
	sqb.Order("id", !asc)
	sqb.Offset(start).Limit(limit)

	return sqb
}

func (sjc *sjc) FindJobChain(name string, asc bool, status ...string) (jc *xjm.JobChain, err error) {
	sqb := sjc.findJobChains(name, 0, 1, asc, status...)
	sql, args := sqb.Build()

	jc = &xjm.JobChain{}
	err = sjc.db.Get(jc, sql, args...)
	if errors.Is(err, sqlx.ErrNoRows) {
		return nil, nil
	}

	return jc, err
}

func (sjc *sjc) FindJobChains(name string, start, limit int, asc bool, status ...string) (jcs []*xjm.JobChain, err error) {
	sqb := sjc.findJobChains(name, start, limit, asc, status...)
	sql, args := sqb.Build()

	err = sjc.db.Select(&jcs, sql, args...)
	if errors.Is(err, sqlx.ErrNoRows) {
		return nil, nil
	}
	return
}

func (sjc *sjc) IterJobChains(it func(*xjm.JobChain) error, name string, start, limit int, asc bool, status ...string) error {
	sqb := sjc.findJobChains(name, start, limit, asc, status...)
	sql, args := sqb.Build()

	rows, err := sjc.db.Queryx(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		jc := &xjm.JobChain{}

		if err := rows.StructScan(jc); err != nil {
			return err
		}

		if err := it(jc); err != nil {
			return err
		}
	}
	return nil
}

func (sjc *sjc) CreateJobChain(name, states string) (int64, error) {
	now := time.Now()

	sqb := sjc.db.Builder()
	sqb.Insert(sjc.tb)
	sqb.Setc("name", name)
	sqb.Setc("status", xjm.JobStatusPending)
	sqb.Setc("states", states)
	sqb.Setc("created_at", now)
	sqb.Setc("updated_at", now)

	if !sjc.db.SupportLastInsertID() {
		sqb.Returns("id")
	}

	sql, args := sqb.Build()
	return sjc.db.Create(sql, args...)
}

func (sjc *sjc) UpdateJobChain(cid int64, status string, states ...string) error {
	if status == "" && len(states) == 0 {
		return nil
	}

	sqb := sjc.db.Builder()

	sqb.Update(sjc.tb)
	if status != "" {
		sqb.Setc("status", status)
	}
	if len(states) > 0 {
		sqb.Setc("states", states[0])
	}
	sqb.Setc("updated_at", time.Now())
	sqb.Where("id = ?", cid)

	sql, args := sqb.Build()

	cnt, err := sjc.db.Update(sql, args...)
	if err != nil {
		return err
	}

	if cnt != 1 {
		return xjm.ErrJobChainMissing
	}
	return nil
}

func (sjc *sjc) DeleteJobChains(cids ...int64) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}

	sqb := sjc.db.Builder()
	sqb.Delete(sjc.tb)
	sqb.In("id", cids)
	sql, args := sqb.Build()

	return sjc.db.Update(sql, args...)
}

func (sjc *sjc) CleanOutdatedJobChains(before time.Time) (int64, error) {
	sqb := sjc.db.Builder()
	sqb.Delete(sjc.tb)
	sqb.Where("updated_at < ?", before)
	sqb.In("status", xjm.JobDoneStatus)
	sql, args := sqb.Build()

	return sjc.db.Update(sql, args...)
}
