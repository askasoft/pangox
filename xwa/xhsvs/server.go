package xhsvs

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/str"
)

var (
	// semaphore channel to limit connections
	semaphore chan struct{}

	// TLLs limited listeners
	TLLs []*netx.LimitedListener

	// TDLs dump listeners
	TDLs []*netx.DumpListener

	// TCPs TCP listeners
	TCPs []net.Listener

	// HTTP http servers
	HSVs []*http.Server
)

type GetCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)

// InitServers initialize TCP listeners and HTTP servers
func InitServers(hh http.Handler, certs ...GetCertificate) error {
	listen := ini.GetString("server", "listen")

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

		tll := netx.NewLimitedListener(tcp, 0)
		tdl := netx.NewDumpListener(tll, "logs")

		hsv := &http.Server{
			Addr:    addr,
			Handler: hh,
		}

		if ssl {
			cert, ok := asg.FindFunc(certs, func(c GetCertificate) bool { return c != nil })
			if !ok {
				return errors.New("xhsvs: nil TLS certificate function")
			}

			hsv.TLSConfig = &tls.Config{
				GetCertificate: cert,
			}
		}

		TCPs = append(TCPs, tcp)
		TLLs = append(TLLs, tll)
		TDLs = append(TDLs, tdl)
		HSVs = append(HSVs, hsv)
	}

	ConfigServers()
	return nil
}

// ConfigServers config http servers
func ConfigServers() {
	maxcon := max(ini.GetInt("server", "maxConnections"), 0)

	if cap(semaphore) != maxcon {
		semaphore = make(chan struct{}, maxcon)
		for _, ttl := range TLLs {
			ttl.Semaphore = semaphore
		}
	}

	for _, tdl := range TDLs {
		tdl.Disable(!ini.GetBool("server", "tcpDump"))
	}

	for _, hsv := range HSVs {
		hsv.ReadHeaderTimeout = ini.GetDuration("server", "httpReadHeaderTimeout", 10*time.Second)
		hsv.ReadTimeout = ini.GetDuration("server", "httpReadTimeout", 120*time.Second)
		hsv.WriteTimeout = ini.GetDuration("server", "httpWriteTimeout", 300*time.Second)
		hsv.IdleTimeout = ini.GetDuration("server", "httpIdleTimeout", 30*time.Second)
	}
}

// ReloadServers reload server configurations
func ReloadServers() error {
	ConfigServers()
	return nil
}

// Serves start serve http servers in go-routines (non-blocking)
func Serves() {
	for i, hsv := range HSVs {
		go serve(hsv, TDLs[i])

		// sleep some time to keep log order
		time.Sleep(10 * time.Millisecond)
	}
}

// Shutdowns gracefully shutdown the http servers with timeout '[server] shutdownTimeout' (default 15 seconds).
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
	timeout := ini.GetDuration("server", "shutdownTimeout", 15*time.Second)
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	log.Infof("HTTP Server %s shutting down in %v ...", hsv.Addr, timeout)

	if err := hsv.Shutdown(ctx); err != nil {
		log.Errorf("HTTP Server %s failed to shutdown: %v", hsv.Addr, err)
	}
}
