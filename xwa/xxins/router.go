package xxins

import (
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa/xtpls"
)

var (
	// XIN global xin engine
	XIN *xin.Engine
)

func InitRouter() {
	XIN = xin.New()

	XIN.HTMLRenderer = xtpls.HTMLRenderer
}

func ConfigRouter() {
	trustedProxies := str.Fields(ini.GetString("server", "httpTrustedProxies"))
	switch len(trustedProxies) {
	case 0:
		trustedProxies = xin.DefaultTrustedProxies
	case 1:
		switch trustedProxies[0] {
		case "*", "anywhere":
			trustedProxies = netx.AnywhereCIDRs
		case "intranet":
			trustedProxies = netx.IntranetCIDRs
		}
	}
	if err := XIN.SetTrustedProxies(trustedProxies); err != nil {
		log.Errorf("invalid setting [server] httpTrustedProxies = %s", str.Join(trustedProxies, " "))
	}

	XIN.TrustedIPHeader = ini.GetString("server", "httpTrustedIPHeader")

	remoteIPHeaders := str.Fields(ini.GetString("server", "httpRemoteIPHeaders"))
	XIN.RemoteIPHeaders = gog.If(len(remoteIPHeaders) > 0, remoteIPHeaders, xin.DefaultRemoteIPHeaders)

	sslProxyHeaders := str.Fields(ini.GetString("server", "httpSSLProxyHeaders"))
	if len(sslProxyHeaders) == 0 {
		XIN.SSLProxyHeaders = xin.DefaultSSLProxyHeaders
	} else {
		hm := make(map[string]string, len(sslProxyHeaders))
		for _, s := range sslProxyHeaders {
			h, v, ok := str.CutByte(s, ':')
			if ok && h != "" {
				hm[h] = v
			}
		}
		XIN.SSLProxyHeaders = gog.If(len(hm) > 0, hm, xin.DefaultSSLProxyHeaders)
	}
}
