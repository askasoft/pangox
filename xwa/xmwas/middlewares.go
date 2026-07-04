package xmwas

import (
	"encoding/json"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/middleware"
	"github.com/askasoft/pangox/xwa"
)

var (
	// XAL global xin access logger
	XAL *middleware.AccessLogger

	// XSL global xin request size limiter
	XSL *middleware.RequestSizeLimiter

	// XRC global xin response compressor
	XRC *middleware.ResponseCompressor

	// XHD global xin http dumper
	XHD *middleware.HTTPDumper

	// XSR global xin https redirector
	XSR *middleware.HTTPSRedirector

	// XLL global xin localizer
	XLL *middleware.Localizer

	// XTP global xin token protector
	XTP *middleware.TokenProtector

	// XRH global xin response header middleware
	XRH *middleware.ResponseHeader

	// XAC global xin origin access controller middleware
	XAC *middleware.OriginAccessController

	// XCC global xin static cache control setter
	XCC *xin.CacheControlSetter
)

func InitMiddlewares() {
	XAL = middleware.NewAccessLogger(nil)
	XSL = middleware.NewRequestSizeLimiter(0)
	XSL.BodyTooLarge = BodyTooLarge
	XRC = middleware.DefaultResponseCompressor()
	XHD = middleware.NewHTTPDumper(log.GetOutputer("XHD", log.LevelTrace))
	XSR = middleware.NewHTTPSRedirector()
	XLL = middleware.NewLocalizer()
	XTP = middleware.NewTokenProtector("")
	XTP.AbortFunc = InvalidToken
	XRH = middleware.NewResponseHeader(nil)
	XAC = middleware.NewOriginAccessController()
	XCC = xin.NewCacheControlSetter()
}

func ConfigMiddlewares() {
	XLL.Locales = xwa.Locales
	XSL.DrainBody = ini.GetBool("router", "httpDrainRequestBody", false)
	XSL.MaxBodySize = ini.GetSize("router", "httpMaxRequestBodySize", 8<<20)

	XRC.Disable(!ini.GetBool("router", "httpGzip"))
	XHD.Disable(!ini.GetBool("router", "httpDump"))
	XSR.Disable(!ini.GetBool("router", "httpsRedirect"))
	XCC.CacheControl = ini.GetString("router", "staticCacheControl", "public, max-age=31536000, immutable")

	XTP.CookiePath = str.IfEmpty(xwa.Base, "/")
	XTP.SetSecret(xwa.Secret)
	XTP.CookieMaxAge = ini.GetDuration("csrf", "cookieMaxAge", time.Hour*24*30)
	XTP.CookieSecure = ini.GetBool("csrf", "cookieSecure", true)
	XTP.SetCookieSameSite(ini.GetString("csrf", "cookieSameSite", "strict"))

	ConfigOriginAccessController(XAC)
	ConfigResponseHeader(XRH)
	ConfigAccessLogger(XAL)
}

func ConfigOriginAccessController(xac *middleware.OriginAccessController) {
	xac.SetAllowOrigins(str.Fields(ini.GetString("router", "accessControlAllowOrigin"))...)
	xac.SetAllowCredentials(ini.GetBool("router", "accessControlAllowCredentials"))
	xac.SetAllowHeaders(ini.GetString("router", "accessControlAllowHeaders"))
	xac.SetAllowMethods(ini.GetString("router", "accessControlAllowMethods"))
	xac.SetExposeHeaders(ini.GetString("router", "accessControlExposeHeaders"))
	xac.SetMaxAge(ini.GetInt("router", "accessControlMaxAge"))
}

func ConfigResponseHeader(xrh *middleware.ResponseHeader) {
	hm := map[string]string{}
	hh := ini.GetString("router", "httpResponseHeader")
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
	alfs := str.Fields(ini.GetString("router", "accessLog"))
	for _, alf := range alfs {
		switch alf {
		case "text":
			alw := middleware.NewAccessLogWriter(
				log.GetOutputer("XAT", log.LevelTrace),
				ini.GetString("router", "accessLogTextFormat", middleware.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := middleware.NewAccessLogWriter(
				log.GetOutputer("XAJ", log.LevelTrace),
				ini.GetString("router", "accessLogJSONFormat", middleware.AccessLogJSONFormat),
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
