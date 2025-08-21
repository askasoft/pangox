package xmsgs

import (
	"io/fs"
	"path/filepath"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tbs"
)

var (
	Root string // current external messages root path
	FS   fs.FS  // internal embedded file system
)

func InitMessages() error {
	root := ini.GetString("app", "messages")

	tb := tbs.NewTextBundles()
	if root != "" {
		log.Infof("Loading messages from '%s'", root)
		if err := tb.Load(root); err != nil {
			return err
		}
	} else if FS != nil {
		log.Info("Loading embedded messages")
		if err := tb.LoadFS(FS, "."); err != nil {
			return err
		}
	}

	Root = root
	tbs.SetDefault(tb)
	return nil
}

func ReloadMessages() bool {
	root := ini.GetString("app", "messages")

	if root != "" {
		log.Infof("Reloading messages from '%s'", root)

		tb := tbs.NewTextBundles()
		if err := tb.Load(root); err != nil {
			log.Errorf("Failed to reload messages from '%s': %v", root, err)
			return false
		}

		Root = root
		tbs.SetDefault(tb)
		return true
	}

	// internal embedded file system
	if root != Root && FS != nil {
		log.Info("Reloading embedded messages")

		tb := tbs.NewTextBundles()
		if err := tb.LoadFS(FS, "."); err != nil {
			log.Errorf("Failed to reload embedded messages: %v", err)
			return false
		}

		Root = root
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

	root := ini.GetString("app", "messages")
	if root == "" || root != Root {
		log.Infof("Skip message reload, no root path set or path changed: '%s' != '%s'", root, Root)
		return false
	}

	// reload on message file change
	log.Infof("Reloading messages from '%s' on [%s] '%s'", root, op, path)

	tb := tbs.NewTextBundles()
	if err := tb.Load(root); err != nil {
		log.Errorf("Failed to reload messages from '%s': %v", root, err)
		return false
	}

	tbs.SetDefault(tb)
	return true
}
