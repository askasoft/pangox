package xtpls

import (
	"fmt"
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
		absdir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("filepath.Abs('%s'): %w", dir, err)
		}

		log.Infof("Loading templates from '%s'", absdir)
		if err := ht.Load(absdir); err != nil {
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

func ReloadTemplates() error {
	dir := ini.GetString("app", "templates")

	if dir != "" {
		absdir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		log.Infof("Reloading templates from '%s'", absdir)

		ht := newHTMLTemplates()
		if err := ht.Load(absdir); err != nil {
			return err
		}

		Dir = dir
		XHT = ht
		return nil
	}

	// internal embedded file system
	// [dir != Dir] means switch from local to embedded
	if dir != Dir && len(FSs) > 0 {
		log.Info("Reloading embedded templates")

		ht := newHTMLTemplates()
		for _, fs := range FSs {
			if err := ht.LoadFS(fs, "."); err != nil {
				return err
			}
		}

		Dir = dir
		XHT = ht
		return nil
	}

	return nil
}

func ReloadTemplatesOnChange(path string, op string) error {
	ext := filepath.Ext(path)
	if !asg.Contains(tpl.HTMLTemplateExtensions, ext) {
		log.Warnf("Skip template reload, unsupported extension: '%s'", ext)
		return nil
	}

	dir := ini.GetString("app", "templates")

	if dir == "" {
		log.Warn("Skip template reload, empty '[app] templates' setting")
		return nil
	}

	if dir != Dir {
		log.Warnf("Skip template reload, '[app] templates' setting changed: '%s' != '%s'", dir, Dir)
		return nil
	}

	absdir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// reload on template file change
	log.Infof("Reloading templates from '%s' on [%s] '%s'", absdir, op, path)

	ht := newHTMLTemplates()
	if err := ht.Load(absdir); err != nil {
		return err
	}

	XHT = ht
	return nil
}
