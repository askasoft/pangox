package sqlxfs

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox/xfs"
)

// sfs implements xfs.XFS interface
type sfs struct {
	db sqlx.Sqlx
	tb string // file table
}

func FS(db sqlx.Sqlx, table string) xfs.XFS {
	return &sfs{db, table}
}

func (sfs *sfs) Open(name string) (fs.File, error) {
	f, err := sfs.FindFile(name)
	if err != nil {
		return nil, err
	}

	hf := &xfs.FSFile{XFS: sfs, File: f}
	return hf, nil
}

// FindFile find a file
func (sfs *sfs) FindFile(id string) (*xfs.File, error) {
	sqb := sfs.db.Builder()
	sqb.Select("id", "name", "ext", "tag", "size", "time")
	sqb.From(sfs.tb).Where("id = ?", id)
	sql, args := sqb.Build()

	f := &xfs.File{}
	err := sfs.db.Get(f, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}

	return f, nil
}

func (sfs *sfs) SaveFile(id string, filename string, filetime time.Time, data []byte, tag ...string) (*xfs.File, error) {
	name := filepath.Base(filename)
	fext := str.ToLower(filepath.Ext(filename))

	fi := &xfs.File{
		ID:   id,
		Name: name,
		Ext:  fext,
		Tag:  asg.First(tag),
		Size: int64(len(data)),
		Time: filetime,
		Data: data,
	}

	sqb := sfs.db.Builder()
	if _, err := sfs.FindFile(id); err == nil {
		sqb.Update(sfs.tb)
		sqb.Setc("name", fi.Name)
		sqb.Setc("ext", fi.Ext)
		sqb.Setc("tag", fi.Tag)
		sqb.Setc("size", fi.Size)
		sqb.Setc("time", fi.Time)
		sqb.Setc("data", fi.Data)
		sqb.Where("id = ?", id)
	} else {
		sqb.Insert(sfs.tb)
		sqb.Setc("id", fi.ID)
		sqb.Setc("name", fi.Name)
		sqb.Setc("ext", fi.Ext)
		sqb.Setc("tag", fi.Tag)
		sqb.Setc("size", fi.Size)
		sqb.Setc("time", fi.Time)
		sqb.Setc("data", fi.Data)
	}
	sql, args := sqb.Build()

	cnt, err := sfs.db.Update(sql, args...)
	if err != nil {
		return fi, err
	}

	if cnt != 1 {
		return fi, fs.ErrNotExist
	}
	return fi, nil
}

func (sfs *sfs) ReadFile(id string) ([]byte, error) {
	sqb := sfs.db.Builder()
	sqb.Select().From(sfs.tb).Where("id = ?", id)
	sql, args := sqb.Build()

	f := &xfs.File{}
	err := sfs.db.Get(f, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}

	return f.Data, nil
}

func (sfs *sfs) CopyFile(src, dst string, tag ...string) error {
	tb := sfs.db.Quote(sfs.tb)

	var args []any

	sql := fmt.Sprintf("INSERT INTO %s (id, name, ext, tag, time, size, data) ", tb)
	if len(tag) == 0 {
		sql += fmt.Sprintf("SELECT ?, name, ext, tag, time, size, data FROM %s WHERE id = ?", tb)
		args = append(args, dst, src)
	} else {
		sql += fmt.Sprintf("SELECT ? AS id, name, ext, ? AS tag, time, size, data FROM %s WHERE id = ?", tb)
		args = append(args, dst, tag[0], src)
	}
	sql = sfs.db.Rebind(sql)

	cnt, err := sfs.db.Update(sql, args...)
	if err != nil {
		return err
	}

	if cnt == 0 {
		return fs.ErrNotExist
	}
	return nil
}

func (sfs *sfs) MoveFile(src, dst string, tag ...string) error {
	sqb := sfs.db.Builder()
	sqb.Update(sfs.tb)
	sqb.Setc("id", dst)
	if len(tag) > 0 {
		sqb.Setc("tag", tag[0])
	}
	sqb.Where("id = ?", src)
	sql, args := sqb.Build()

	cnt, err := sfs.db.Update(sql, args...)
	if err != nil {
		return err
	}

	if cnt == 0 {
		return fs.ErrNotExist
	}
	return nil
}

func (sfs *sfs) DeleteFile(id string) error {
	sqb := sfs.db.Builder()
	sqb.Delete(sfs.tb).Where("id = ?", id)
	sql, args := sqb.Build()

	_, err := sfs.db.Exec(sql, args...)
	return err
}

func (sfs *sfs) DeleteFiles(ids ...string) (int64, error) {
	sql, args := sqx.In("id", ids)
	return sfs.DeleteWhere(sql, args...)
}

func (sfs *sfs) DeletePrefix(prefix string) (int64, error) {
	return sfs.DeleteWhere("id LIKE ?", sqx.StartsLike(prefix))
}

func (sfs *sfs) DeleteTagged(tag string) (int64, error) {
	return sfs.DeleteWhere("tag = ?", tag)
}

func (sfs *sfs) DeleteBefore(before time.Time) (int64, error) {
	return sfs.DeleteWhere("time < ?", before)
}

func (sfs *sfs) DeletePrefixBefore(prefix string, before time.Time) (int64, error) {
	return sfs.DeleteWhere("id LIKE ? AND time < ?", sqx.StartsLike(prefix), before)
}

func (sfs *sfs) DeleteTaggedBefore(tag string, before time.Time) (int64, error) {
	return sfs.DeleteWhere("tag = ? AND time < ?", tag, before)
}

func (sfs *sfs) DeleteWhere(where string, args ...any) (int64, error) {
	sql := sfs.db.Rebind("DELETE FROM " + sfs.db.Quote(sfs.tb) + " WHERE " + where)
	return sfs.db.Update(sql, args...)
}

// DeleteAll use "DELETE FROM files" to delete all files
func (sfs *sfs) DeleteAll() (int64, error) {
	return sfs.db.Update("DELETE FROM " + sfs.db.Quote(sfs.tb))
}

// Truncate use "TRUNCATE TABLE files" to truncate files
func (sfs *sfs) Truncate() error {
	_, err := sfs.db.Exec("TRUNCATE TABLE " + sfs.db.Quote(sfs.tb))
	return err
}
