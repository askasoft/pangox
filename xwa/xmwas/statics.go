package xmwas

import (
	"io/fs"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa"
)

type dynafs struct {
	FS *fs.FS
}

func (dfs dynafs) Open(name string) (fs.File, error) {
	return (*dfs.FS).Open(name)
}

func AddStaticsHandlers(rg *xin.RouterGroup, statics map[string]fs.FS) {
	for path, fs := range statics {
		xin.StaticFS(rg, path, xin.FS(fsu.FixedModTimeFS(fs, xwa.BuildTime)), XCC.Handle)
	}
}

func AddDynamicFolderHandlers(rg *xin.RouterGroup, dfs *fs.FS, suffixs ...string) {
	wfs := fsu.FixedModTimeFS(dynafs{dfs}, xwa.BuildTime)

	sfx := str.NonEmpty(suffixs...)
	if sfx != "" {
		sfx = "/" + sfx
	}

	des := gog.Must(fs.ReadDir(*dfs, "."))
	for _, de := range des {
		path := de.Name() + sfx
		if de.IsDir() {
			xin.StaticFS(rg, path, xin.FS(fsu.MustSubFS(wfs, de.Name())), XCC.Handle)
		}
	}
}

func AddDynamicFileHandlers(rg *xin.RouterGroup, dfs *fs.FS) {
	wfs := fsu.FixedModTimeFS(dynafs{dfs}, xwa.BuildTime)
	hfs := xin.FS(wfs)

	des := gog.Must(fs.ReadDir(*dfs, "."))
	for _, de := range des {
		if !de.IsDir() {
			xin.StaticFSFile(rg, de.Name(), hfs, de.Name(), XCC.Handle)
		}
	}
}
