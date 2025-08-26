package xwa

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ids/npid"
	"github.com/askasoft/pango/ids/snowflake"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
)

var (
	// LogConfigFile log config file
	LogConfigFile = "conf/log.ini"

	// AppConfigFile app config file
	AppConfigFiles = []string{"conf/app.ini", "conf/env.ini"}

	// CFG global ini map
	CFG map[string]map[string]string

	// Base web context path
	Base string

	// Domain site domain
	Domain string

	// Secret secret string used for token protection
	Secret string

	// Locales supported languages
	Locales []string
)

// inject by go build
var (
	// Version app version inject by go build
	Version string

	// Revision app revision inject by go build
	Revision string

	// Buildtime app build time "2006-01-02T15:04:05Z" inject by go build
	Buildtime string
)

var (
	// BuildTime app build time
	BuildTime time.Time

	// StartupTime app start time
	StartupTime = time.Now()

	// InstanceID app instance ID
	InstanceID = npid.New(10, 0)

	// Sequencer app snowflake ID generator
	Sequencer = snowflake.NewNode(InstanceID)
)

// init built-in variables on debug
func init() {
	if Buildtime == "" {
		BuildTime = StartupTime
	} else {
		BuildTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", Buildtime, time.UTC)
	}

	if Revision == "" {
		Revision = fmt.Sprintf("%x", BuildTime.Unix())
	}

	if Version == "" {
		Version = "0.0.0"
	}
}

func Versions() string {
	return fmt.Sprintf("%s.%s (%s) [%s %s/%s]", Version, Revision, BuildTime.Local(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func InitLogs() error {
	if err := log.Config(LogConfigFile); err != nil {
		return err
	}

	log.SetProp("VERSION", Version)
	log.SetProp("REVISION", Revision)
	golog.SetOutput(log.GetOutputer("std", log.LevelInfo, 2))

	dir, _ := filepath.Abs(".")
	log.Info("Initializing ...")
	log.Infof("Runtime:    %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Infof("Version:    %s.%s", Version, Revision)
	log.Infof("BuildTime:  %s", BuildTime.Local())
	log.Infof("Directory:  %s", dir)
	log.Infof("ProcessID:  %d", os.Getpid())
	log.Infof("InstanceID: 0x%x", InstanceID)

	return nil
}

func ReloadLogs(op string) error {
	log.Infof("Reloading log '%s' [%s]", LogConfigFile, op)

	if err := log.Config(LogConfigFile); err != nil {
		return fmt.Errorf("invalid log configuration '%s': %v", LogConfigFile, err)
	}

	return nil
}

func InitConfigs() error {
	cfg, err := LoadConfigs()
	if err != nil {
		return err
	}

	ini.SetDefault(cfg)

	CFG = ini.StringMap()
	Base = ini.GetString("server", "prefix")
	Domain = ini.GetString("server", "domain")
	Secret = ini.GetString("app", "secret", "~ pangoxsecret ~")
	Locales = str.FieldsAny(ini.GetString("app", "locales"), ",; ")

	return nil
}

func LoadConfigs() (*ini.Ini, error) {
	c := ini.NewIni()

	for i, f := range AppConfigFiles {
		if i > 0 && fsu.FileExists(f) != nil {
			continue
		}

		log.Infof("Loading config: '%s'", f)
		if err := c.LoadFile(f); err != nil {
			return nil, fmt.Errorf("invalid config file '%s': %w", f, err)
		}
	}

	return c, nil
}

func MakeFileID(prefix, name string) string {
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + num.Ltoa(Sequencer.NextID().Int64()) + "/"

	_, name = filepath.Split(name)
	name = str.RemoveAny(name, `\/:*?"<>|`)

	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}
