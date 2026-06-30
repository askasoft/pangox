package xmwas

import (
	"io/fs"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa"
)

func AddStaticsHandlers(rg *xin.RouterGroup, statics map[string]fs.FS) {
	for path, fs := range statics {
		xin.StaticFS(rg, path, xin.FS(fsu.FixedModTimeFS(fs, xwa.BuildTime)), XCC.Handle)
	}
}

func AddStaticSubFolderHandlers(rg *xin.RouterGroup, rfs fs.FS, suffixs ...string) error {
	wfs := fsu.FixedModTimeFS(rfs, xwa.BuildTime)

	sfx := asg.First(suffixs)
	if sfx != "" {
		sfx = "/" + sfx
	}

	des, err := fs.ReadDir(rfs, ".")
	if err != nil {
		return err
	}

	for _, de := range des {
		if de.IsDir() {
			fsub, err := fs.Sub(wfs, de.Name())
			if err != nil {
				return err
			}

			xin.StaticFS(rg, de.Name()+sfx, xin.FS(fsub), XCC.Handle)
		}
	}
	return nil
}

func AddStaticSubFileHandlers(rg *xin.RouterGroup, rfs fs.FS) error {
	wfs := fsu.FixedModTimeFS(rfs, xwa.BuildTime)
	hfs := xin.FS(wfs)

	des, err := fs.ReadDir(rfs, ".")
	if err != nil {
		return err
	}

	for _, de := range des {
		if !de.IsDir() {
			xin.StaticFSFile(rg, de.Name(), hfs, de.Name(), XCC.Handle)
		}
	}
	return nil
}
