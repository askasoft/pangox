package xmws

import (
	"encoding/json"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin/middleware"
	"github.com/askasoft/pangox/xwa"
)

func ConfigResponseHeader(xrh *middleware.ResponseHeader) {
	hm := map[string]string{}
	hh := ini.GetString("server", "httpResponseHeader")
	if hh == "" {
		xrh.Header = hm
	} else {
		err := json.Unmarshal(str.UnsafeBytes(hh), &hm)
		if err == nil {
			sr := str.NewReplacer(
				"{{VERSION}}", xwa.Version,
				"{{REVISION}}", xwa.Revision,
				"{{BuildTime}}", xwa.BuildTime.Format(time.RFC3339),
			)
			for k, v := range hm {
				hm[k] = sr.Replace(v)
			}
			xrh.Header = hm
		} else {
			log.Errorf("Invalid httpResponseHeader '%s': %v", hh, err)
		}
	}
}

func ConfigAccessLogger(xal *middleware.AccessLogger) {
	alws := []middleware.AccessLogWriter{}
	alfs := str.Fields(ini.GetString("server", "accessLog"))
	for _, alf := range alfs {
		switch alf {
		case "text":
			alw := middleware.NewAccessLogWriter(
				log.GetOutputer("XAT", log.LevelTrace),
				ini.GetString("server", "accessLogTextFormat", middleware.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := middleware.NewAccessLogWriter(
				log.GetOutputer("XAJ", log.LevelTrace),
				ini.GetString("server", "accessLogJSONFormat", middleware.AccessLogJSONFormat),
			)
			alws = append(alws, alw)
		default:
			log.Warnf("Invalid accessLog setting: %s", alf)
		}
	}

	switch len(alws) {
	case 0:
		xal.SetWriter(nil)
	case 1:
		xal.SetWriter(alws[0])
	default:
		xal.SetWriter(middleware.NewAccessLogMultiWriter(alws...))
	}
}
