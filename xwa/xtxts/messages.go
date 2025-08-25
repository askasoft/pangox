package xtxts

import (
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
		log.Infof("Loading messages from '%s'", dir)
		if err := tb.Load(dir); err != nil {
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

func ReloadMessages() bool {
	dir := ini.GetString("app", "messages")

	if dir != "" {
		log.Infof("Reloading messages from '%s'", dir)

		tb := tbs.NewTextBundles()
		if err := tb.Load(dir); err != nil {
			log.Errorf("Failed to reload messages from '%s': %v", dir, err)
			return false
		}

		Dir = dir
		tbs.SetDefault(tb)
		return true
	}

	// internal embedded file system
	if dir != Dir && len(FSs) > 0 {
		log.Info("Reloading embedded messages")

		tb := tbs.NewTextBundles()
		for _, fs := range FSs {
			if err := tb.LoadFS(fs, "."); err != nil {
				log.Errorf("Failed to reload embedded messages: %v", err)
				return false
			}
		}

		Dir = dir
		tbs.SetDefault(tb)
		return true
	}

	return false
}

func ReloadMessagesOnChange(path string, op string) bool {
	ext := filepath.Ext(path)
	if !asg.Contains(tbs.Default().Extensions, ext) {
		log.Infof("Skip message reload, unsupported extension: '%s'", ext)
		return false
	}

	dir := ini.GetString("app", "messages")
	if dir == "" || dir != Dir {
		log.Infof("Skip message reload, no dir path set or path changed: '%s' != '%s'", dir, Dir)
		return false
	}

	// reload on message file change
	log.Infof("Reloading messages from '%s' on [%s] '%s'", dir, op, path)

	tb := tbs.NewTextBundles()
	if err := tb.Load(dir); err != nil {
		log.Errorf("Failed to reload messages from '%s': %v", dir, err)
		return false
	}

	tbs.SetDefault(tb)
	return true
}
