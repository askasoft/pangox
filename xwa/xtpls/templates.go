package xtpls

import (
	"io/fs"
	"path/filepath"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tpl"
	"github.com/askasoft/pango/xin/render"
	"github.com/askasoft/pango/xin/taglib"
)

var (
	XHT   tpl.Templates // global html templates
	Funcs tpl.FuncMap   // custom template functions
	Root  string        // current external templates root path
	FS    fs.FS         // internal embedded file system
)

func HTMLRenderer(locale, name string, data any) render.Render {
	return render.HTMLRender{
		Templates: XHT,
		Locale:    locale,
		Name:      name,
		Data:      data,
	}
}

// Functions default utility functions for template
func Functions() tpl.FuncMap {
	fm := tpl.Functions()
	fm.Copy(taglib.Functions())
	fm.Copy(Funcs)
	return fm
}

func newHTMLTemplates() tpl.Templates {
	ht := tpl.NewHTMLTemplates()

	fm := Functions()
	ht.Funcs(fm)

	return ht
}

func InitTemplates() error {
	root := ini.GetString("app", "templates")

	ht := newHTMLTemplates()
	if root != "" {
		log.Infof("Loading templates from '%s'", root)
		if err := ht.Load(root); err != nil {
			return err
		}
	} else if FS != nil {
		log.Info("Loading embedded templates")
		if err := ht.LoadFS(FS, "."); err != nil {
			return err
		}
	}

	XHT = ht
	Root = root
	return nil
}

func ReloadTemplates() bool {
	root := ini.GetString("app", "templates")

	if root != "" {
		log.Infof("Reloading templates from '%s'", root)

		ht := newHTMLTemplates()
		if err := ht.Load(root); err != nil {
			log.Errorf("Failed to reload templates from '%s': %v", root, err)
			return false
		}

		Root = root
		XHT = ht
		return true
	}

	// internal embedded file system
	if root != Root && FS != nil {
		log.Info("Reloading embedded templates")

		ht := newHTMLTemplates()
		if err := ht.LoadFS(FS, "."); err != nil {
			log.Errorf("Failed to reload embedded templates: %v", err)
			return false
		}

		Root = root
		XHT = ht
		return true
	}

	return false
}

func ReloadTemplatesOnChange(path string, op string) bool {
	ext := filepath.Ext(path)
	if !asg.Contains(tpl.HTMLTemplateExtensions, ext) {
		log.Infof("Skip template reload, unsupported extension: '%s'", ext)
		return false
	}

	root := ini.GetString("app", "templates")

	if root == "" || root != Root {
		log.Infof("Skip template reload, no root path set or path changed: '%s' != '%s'", root, Root)
		return false
	}

	// reload on template file change
	log.Infof("Reloading templates from '%s' on [%s] '%s'", root, op, path)

	ht := newHTMLTemplates()
	if err := ht.Load(root); err != nil {
		log.Errorf("Failed to reload templates from '%s': %v", root, err)
		return false
	}

	XHT = ht
	return true
}
