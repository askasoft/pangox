package xosms

import (
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/oss/osm"
)

func InitStatsMonitor() {
	sm := osm.Default()

	sm.Logger = log.GetLogger("OSM")

	RunStatsMonitor()
}

func RunStatsMonitor() {
	sm := osm.Default()

	sm.Interval = ini.GetDuration("monitor", "interval")
	sm.DiskFree = ini.GetSize("monitor", "diskFree", num.GB)
	sm.DiskCount = ini.GetInt("monitor", "diskCount", 5)
	sm.CPUUsage = ini.GetFloat("monitor", "cpuUsage", 0.9)
	sm.CPUCount = ini.GetInt("monitor", "cpuCount", 5)
	sm.MemUsage = ini.GetFloat("monitor", "memUsage", 0.9)
	sm.MemCount = ini.GetInt("monitor", "memCount", 5)

	if sm.Interval > 0 {
		sm.Start()
		return
	}

	sm.Stop()
}
