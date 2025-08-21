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

	// Locales supported languages
	Locales []string
)

// inject by go build
var (
	// version app version
	version string

	// revision app revision
	revision string

	// buildtime app build time
	buildtime string
)

var (
	// buildTime app build time
	buildTime time.Time

	// startupTime app start time
	startupTime = time.Now()

	// instanceID app instance ID
	instanceID = npid.New(10, 0)

	// sequencer app snowflake ID generator
	sequencer = snowflake.NewNode(instanceID)
)

// init built-in variables on debug
func init() {
	if version == "" {
		version = "0"
	}

	if revision == "" {
		revision = fmt.Sprintf("%x", startupTime.Unix())
	}

	if buildtime == "" {
		buildTime = startupTime
	} else {
		buildTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", buildtime, time.UTC)
	}
}

func Version() string {
	return version
}

func Revision() string {
	return revision
}

func Versions() string {
	return fmt.Sprintf("%s.%s (%s) [%s %s/%s]", version, revision, buildTime.Local(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func BuildTime() time.Time {
	return buildTime
}

func StartupTime() time.Time {
	return startupTime
}

func InstanceID() int64 {
	return instanceID
}

func Sequencer() *snowflake.Node {
	return sequencer
}

func Exit(code int) {
	log.Close()
	os.Exit(code)
}

func InitLogs() error {
	if err := log.Config(LogConfigFile); err != nil {
		return err
	}

	log.SetProp("VERSION", Version())
	log.SetProp("REVISION", Revision())
	golog.SetOutput(log.GetOutputer("std", log.LevelInfo, 2))

	dir, _ := filepath.Abs(".")
	log.Info("Initializing ...")
	log.Infof("Version:    %s.%s", Version(), Revision())
	log.Infof("BuildTime:  %s", BuildTime().Local())
	log.Infof("Runtime:    %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Infof("Directory:  %s", dir)
	log.Infof("ProcessID:  %d", os.Getpid())
	log.Infof("InstanceID: 0x%x", InstanceID())

	return nil
}

func ReloadLogs(op string) {
	log.Infof("Reloading log '%s' [%s]", LogConfigFile, op)

	err := log.Config(LogConfigFile)
	if err != nil {
		log.Errorf("Failed to reload log '%s': %v", LogConfigFile, err)
	}
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
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + num.Ltoa(sequencer.NextID().Int64()) + "/"

	_, name = filepath.Split(name)
	name = str.RemoveAny(name, `\/:*?"<>|`)

	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}
