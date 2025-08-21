package xfsws

import (
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox/xwa"
)

var (
	ReloadLogs      func(string, fsw.Op)
	ReloadConfigs   func(string, fsw.Op)
	ReloadMessages  func(string, fsw.Op)
	ReloadTemplates func(string, fsw.Op)
)

// InitFileWatch initialize file watch
func InitFileWatch() error {
	fsw.Default().Logger = log.GetLogger("FSW")

	if ReloadLogs != nil {
		if err := fsw.Add(xwa.LogConfigFile, fsw.OpWrite, ReloadLogs); err != nil {
			return err
		}
	}

	if ReloadConfigs != nil {
		for _, f := range xwa.AppConfigFiles {
			if fsu.FileExists(f) == nil {
				if err := fsw.Add(f, fsw.OpWrite, ReloadConfigs); err != nil {
					return err
				}
			}
		}
	}

	if ReloadMessages != nil {
		if msgPath := ini.GetString("app", "messages"); msgPath != "" {
			if err := fsw.AddRecursive(msgPath, fsw.OpModifies, ReloadMessages); err != nil {
				return err
			}
		}
	}

	if ReloadTemplates != nil {
		if tplPath := ini.GetString("app", "templates"); tplPath != "" {
			if err := fsw.AddRecursive(tplPath, fsw.OpModifies, ReloadTemplates); err != nil {
				return err
			}
		}
	}

	return RunFileWatch()
}

func RunFileWatch() error {
	if ini.GetBool("app", "reloadable") {
		return fsw.Start()
	}

	return fsw.Stop()
}
