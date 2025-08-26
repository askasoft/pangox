package xhsvs

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/str"
)

var (
	// TCPs TCP listeners
	TCPs []*netx.DumpListener

	// HTTP http servers
	HSVs []*http.Server
)

type GetCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)

// InitServers initialize TCP listeners and HTTP servers
func InitServers(hh http.Handler, cert GetCertificate) error {
	listen := ini.GetString("server", "listen", ":6060")

	var semaphore chan struct{}
	maxcon := ini.GetInt("server", "maxConnections")
	if maxcon > 0 {
		semaphore = make(chan struct{}, maxcon)
	}

	for _, addr := range str.Fields(listen) {
		log.Infof("Listening %s ...", addr)

		ssl := str.EndsWithByte(addr, 's')
		if ssl {
			addr = addr[:len(addr)-1]
		}

		tcp, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		if maxcon > 0 {
			tcp = netx.NewLimitListener(tcp, semaphore)
		}

		tcpd := netx.NewDumpListener(tcp, "logs")

		hsv := &http.Server{
			Addr:    addr,
			Handler: hh,
		}

		if ssl {
			if cert == nil {
				return errors.New("xhsvs: nil TLS certificate function")
			}

			hsv.TLSConfig = &tls.Config{
				GetCertificate: cert,
			}
		}
		TCPs = append(TCPs, tcpd)
		HSVs = append(HSVs, hsv)
	}

	ConfigServers()
	return nil
}

// ConfigServers config http servers
func ConfigServers() {
	for _, tcpd := range TCPs {
		tcpd.Disable(!ini.GetBool("server", "tcpDump"))
	}

	for _, hsv := range HSVs {
		hsv.ReadHeaderTimeout = ini.GetDuration("server", "httpReadHeaderTimeout", 10*time.Second)
		hsv.ReadTimeout = ini.GetDuration("server", "httpReadTimeout", 120*time.Second)
		hsv.WriteTimeout = ini.GetDuration("server", "httpWriteTimeout", 300*time.Second)
		hsv.IdleTimeout = ini.GetDuration("server", "httpIdleTimeout", 30*time.Second)
	}
}

// Serves start serve http servers in go-routines (non-blocking)
func Serves() {
	for i, hsv := range HSVs {
		tcp := TCPs[i]
		go serve(hsv, tcp)
		time.Sleep(time.Millisecond)
	}
}

// Shutdowns gracefully shutdown the http servers with timeout '[server] shutdownTimeout' (defautl 5 seconds).
func Shutdowns() {
	// shutdown http servers
	var wg sync.WaitGroup
	for _, hsv := range HSVs {
		wg.Add(1)
		go shutdown(hsv, &wg)
	}
	wg.Wait()
}

func serve(hsv *http.Server, tcp net.Listener) {
	log.Infof("HTTP Serving %s ...", hsv.Addr)

	if hsv.TLSConfig != nil {
		tcp = tls.NewListener(tcp, hsv.TLSConfig)
	}

	if err := hsv.Serve(tcp); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("HTTP Server %s closed", hsv.Addr)
		} else {
			log.Fatalf(99, "HTTP.Serve(%s) failed: %v", hsv.Addr, err)
		}
	}
}

func shutdown(hsv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()

	// The context is used to inform the server it has some seconds to finish
	// the request it is currently handling
	timeout := ini.GetDuration("server", "shutdownTimeout", 5*time.Second)
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	log.Infof("HTTP Server %s shutting down in %v ...", hsv.Addr, timeout)

	if err := hsv.Shutdown(ctx); err != nil {
		log.Errorf("HTTP Server %s failed to shutdown: %v", hsv.Addr, err)
	}
}
