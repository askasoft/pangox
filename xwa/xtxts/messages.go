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
	return loadMessages("Loading")
}

func ReloadMessages() error {
	return loadMessages("Reloading")
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

	log.Infof("Reloading messages on [%s] '%s'", op, path)
	return ReloadMessages()
}

func loadMessages(action string) error {
	tb := tbs.NewTextBundles()
	if len(FSs) > 0 {
		log.Infof("%s embedded messages", action)
		if err := tb.LoadFS(FSs...); err != nil {
			return err
		}
	}

	dir := ini.GetString("app", "messages")

	if dir != "" {
		absdir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("filepath.Abs('%s'): %w", dir, err)
		}

		log.Infof("%s external messages: '%s'", action, absdir)
		if err := tb.Load(absdir); err != nil {
			return err
		}
	}

	Dir = dir
	tbs.SetDefault(tb)
	return nil
}
