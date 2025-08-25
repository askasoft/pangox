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
	Dir   string        // current local templates directory
	FSs   []fs.FS       // internal embedded templates file systems
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
	dir := ini.GetString("app", "templates")

	ht := newHTMLTemplates()
	if dir != "" {
		log.Infof("Loading templates from '%s'", dir)
		if err := ht.Load(dir); err != nil {
			return err
		}
	} else if len(FSs) > 0 {
		log.Info("Loading embedded templates")
		for _, fs := range FSs {
			if err := ht.LoadFS(fs, "."); err != nil {
				return err
			}
		}
	}

	XHT = ht
	Dir = dir
	return nil
}

func ReloadTemplates() bool {
	dir := ini.GetString("app", "templates")

	if dir != "" {
		log.Infof("Reloading templates from '%s'", dir)

		ht := newHTMLTemplates()
		if err := ht.Load(dir); err != nil {
			log.Errorf("Failed to reload templates from '%s': %v", dir, err)
			return false
		}

		Dir = dir
		XHT = ht
		return true
	}

	// internal embedded file system
	if dir != Dir && len(FSs) > 0 {
		log.Info("Reloading embedded templates")

		ht := newHTMLTemplates()
		for _, fs := range FSs {
			if err := ht.LoadFS(fs, "."); err != nil {
				log.Errorf("Failed to reload embedded templates: %v", err)
				return false
			}
		}

		Dir = dir
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

	dir := ini.GetString("app", "templates")

	if dir == "" || dir != Dir {
		log.Infof("Skip template reload, no dir path set or path changed: '%s' != '%s'", dir, Dir)
		return false
	}

	// reload on template file change
	log.Infof("Reloading templates from '%s' on [%s] '%s'", dir, op, path)

	ht := newHTMLTemplates()
	if err := ht.Load(dir); err != nil {
		log.Errorf("Failed to reload templates from '%s': %v", dir, err)
		return false
	}

	XHT = ht
	return true
}
