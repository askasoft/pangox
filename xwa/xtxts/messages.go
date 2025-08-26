package xtxts

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tbs"
)

var (
	Dir string  // current local messages folder
	FSs []fs.FS // internal embedded message file systems
)

func InitMessages() error {
	dir := ini.GetString("app", "messages")

	tb := tbs.NewTextBundles()
	if dir != "" {
		absdir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("filepath.Abs('%s'): %w", dir, err)
		}

		log.Infof("Loading messages from '%s'", absdir)
		if err := tb.Load(absdir); err != nil {
			return err
		}
	} else if len(FSs) > 0 {
		log.Info("Loading embedded messages")
		for _, fs := range FSs {
			if err := tb.LoadFS(fs, "."); err != nil {
				return err
			}
		}
	}

	Dir = dir
	tbs.SetDefault(tb)
	return nil
}

func ReloadMessages() error {
	dir := ini.GetString("app", "messages")

	if dir != "" {
		absdir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		log.Infof("Reloading messages from '%s'", absdir)

		tb := tbs.NewTextBundles()
		if err := tb.Load(absdir); err != nil {
			return err
		}

		Dir = dir
		tbs.SetDefault(tb)
		return nil
	}

	// internal embedded file system
	// [dir != Dir] means switch from local to embedded
	if dir != Dir && len(FSs) > 0 {
		log.Info("Reloading embedded messages")

		tb := tbs.NewTextBundles()
		for _, fs := range FSs {
			if err := tb.LoadFS(fs, "."); err != nil {
				return err
			}
		}

		Dir = dir
		tbs.SetDefault(tb)
		return nil
	}

	return nil
}

func ReloadMessagesOnChange(path string, op string) error {
	ext := filepath.Ext(path)
	if !asg.Contains(tbs.Default().Extensions, ext) {
		log.Warnf("Skip message reload, unsupported extension: '%s'", ext)
		return nil
	}

	dir := ini.GetString("app", "messages")
	if dir == "" {
		log.Warn("Skip message reload, empty '[app] messages' setting")
		return nil
	}

	if dir != Dir {
		log.Warnf("Skip message reload, '[app] messages' setting changed: '%s' != '%s'", dir, Dir)
		return nil
	}

	absdir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// reload on message file change
	log.Infof("Reloading messages from '%s' on [%s] '%s'", absdir, op, path)

	tb := tbs.NewTextBundles()
	if err := tb.Load(absdir); err != nil {
		return err
	}

	tbs.SetDefault(tb)
	return nil
}
