package xfsws

import (
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox/xwa"
)

var (
	ReloadLogsOnChange func(string, fsw.Op)
	ReloadCfgsOnChange func(string, fsw.Op)
	ReloadMsgsOnChange func(string, fsw.Op)
	ReloadTplsOnChange func(string, fsw.Op)
)

// InitFileWatch initialize file watch
func InitFileWatch() error {
	fsw.Default().Logger = log.GetLogger("FSW")

	if !ini.GetBool("app", "reloadable") {
		return nil
	}

	if err := configFileWatch(); err != nil {
		return err
	}

	return fsw.Start()
}

func configFileWatch() error {
	if ReloadLogsOnChange != nil {
		if err := fsw.Add(xwa.LogConfigFile, fsw.OpWrite, ReloadLogsOnChange); err != nil {
			return err
		}
	}

	if ReloadCfgsOnChange != nil {
		for _, f := range xwa.AppConfigFiles {
			if fsu.FileExists(f) == nil {
				if err := fsw.Add(f, fsw.OpWrite, ReloadCfgsOnChange); err != nil {
					return err
				}
			}
		}
	}

	if ReloadMsgsOnChange != nil {
		if msgPath := ini.GetString("app", "messages"); msgPath != "" {
			if err := fsw.AddRecursive(msgPath, fsw.OpModifies, ReloadMsgsOnChange); err != nil {
				return err
			}
		}
	}

	if ReloadTplsOnChange != nil {
		if tplPath := ini.GetString("app", "templates"); tplPath != "" {
			if err := fsw.AddRecursive(tplPath, fsw.OpModifies, ReloadTplsOnChange); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReloadFileWatch() error {
	if err := fsw.Close(); err != nil {
		return err
	}

	if ini.GetBool("app", "reloadable") {
		if err := configFileWatch(); err != nil {
			return err
		}

		if err := fsw.Start(); err != nil {
			return err
		}
	}

	return nil
}

func CloseFileWatch() error {
	return fsw.Close()
}
