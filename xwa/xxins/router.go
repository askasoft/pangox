package xxins

import (
	"fmt"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa/xtpls"
)

var xins = map[string]*xin.Engine{}

func Router(id ...string) *xin.Engine {
	return xins[asg.First(id)]
}

func InitRouter() {
	InitRouters("")
}

func InitRouters(ids ...string) {
	for _, id := range ids {
		r := xin.New()
		r.HTMLRenderer = xtpls.HTMLRenderer
		xins[id] = r
	}
}

func ConfigRouter() {
	ConfigRouters("")
}

func ConfigRouters(ids ...string) {
	sec := ini.GetSection("router")

	for _, id := range ids {
		r := xins[id]
		if r == nil {
			panic(fmt.Errorf("xxins: invalid router '%s'", id))
		}

		trustedProxies := str.Fields(sec.GetString("httpTrustedProxies"))
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
		if err := r.SetTrustedProxies(trustedProxies); err != nil {
			log.Errorf("invalid setting [server] httpTrustedProxies = %s", str.Join(trustedProxies, " "))
		}

		r.TrustedIPHeader = sec.GetString("httpTrustedIPHeader")

		remoteIPHeaders := str.Fields(sec.GetString("httpRemoteIPHeaders"))
		r.RemoteIPHeaders = gog.If(len(remoteIPHeaders) > 0, remoteIPHeaders, xin.DefaultRemoteIPHeaders)

		sslProxyHeaders := str.Fields(sec.GetString("httpSSLProxyHeaders"))
		if len(sslProxyHeaders) == 0 {
			r.SSLProxyHeaders = xin.DefaultSSLProxyHeaders
		} else {
			hm := make(map[string]string, len(sslProxyHeaders))
			for _, s := range sslProxyHeaders {
				h, v, ok := str.CutByte(s, ':')
				if ok && h != "" {
					hm[h] = v
				}
			}
			r.SSLProxyHeaders = gog.If(len(hm) > 0, hm, xin.DefaultSSLProxyHeaders)
		}

	}
}
