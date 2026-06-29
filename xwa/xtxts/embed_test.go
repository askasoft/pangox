package xtxts

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/str"
)

func TestEmbedFS(t *testing.T) {
	fmt.Println("------------------------")
	err := fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path, d.IsDir())
		if d.IsDir() {
			return nil
		}

		i := ini.NewIni()
		if err := i.LoadFileFS(FS, path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		for _, s := range i.Sections() {
			for k, v := range s.StringMap() {
				if str.ContainsRune(k, '\u200b') {
					return fmt.Errorf("%s: [%s] %s contains \\u200b", path, s.Name(), k)
				}
				if str.ContainsRune(v, '\u200b') {
					fmt.Print(str.RemoveRune(v, '\u200b'))
					return fmt.Errorf("%s: [%s] %s 's value contains \\u200b", path, s.Name(), k)
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Println("------------------------")
}
