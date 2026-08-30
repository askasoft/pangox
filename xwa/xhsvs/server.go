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
	"github.com/askasoft/pangox/xwa/xcert"
)

var servers = map[string]*Server{}

type Server struct {
	// ID config id
	ID string

	// Semaphore channel to limit connections
	Semaphore chan struct{}

	// TLLs limited listeners
	TLLs []*netx.LimitedListener

	// TDLs dump listeners
	TDLs []*netx.DumpListener

	// TCPs TCP listeners
	TCPs []net.Listener

	// HTTP http servers
	HSVs []*http.Server
}

func (s *Server) key() string {
	key := "server"
	if s.ID != "" {
		key += "." + s.ID
	}
	return key
}

func (s *Server) Init(hh http.Handler) error {
	sec := ini.GetSection(s.key())

	listen := sec.GetString("listen")
	dumpdir := sec.GetString("tcpDumpDir", "logs")

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
		tdl := netx.NewDumpListener(tll, dumpdir)

		hsv := &http.Server{
			Addr:    addr,
			Handler: hh,
		}

		if ssl {
			hsv.TLSConfig = xcert.TLSConfig
		}

		s.TCPs = append(s.TCPs, tcp)
		s.TLLs = append(s.TLLs, tll)
		s.TDLs = append(s.TDLs, tdl)
		s.HSVs = append(s.HSVs, hsv)
	}

	s.Config()

	return nil
}

// Config config http server
func (s *Server) Config() {
	sec := ini.GetSection(s.key())

	maxcon := max(sec.GetInt("maxConnections"), 0)

	if cap(s.Semaphore) != maxcon {
		s.Semaphore = make(chan struct{}, maxcon)
		for _, ttl := range s.TLLs {
			ttl.Semaphore = s.Semaphore
		}
	}

	for _, tdl := range s.TDLs {
		tdl.Disable(!sec.GetBool("tcpDump"))
		tdl.Path = sec.GetString("tcpDumpDir", "logs")
	}

	for _, hsv := range s.HSVs {
		hsv.ReadHeaderTimeout = sec.GetDuration("httpReadHeaderTimeout", 10*time.Second)
		hsv.ReadTimeout = sec.GetDuration("httpReadTimeout", 120*time.Second)
		hsv.WriteTimeout = sec.GetDuration("httpWriteTimeout", 300*time.Second)
		hsv.IdleTimeout = sec.GetDuration("httpIdleTimeout", 30*time.Second)
	}
}

// Serve start serve http servers in go-routines (non-blocking)
func (s *Server) Serve() {
	for i, hsv := range s.HSVs {
		go s.serve(hsv, s.TDLs[i])

		// sleep some time to keep log order
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Server) serve(hsv *http.Server, tcp net.Listener) {
	if hsv.TLSConfig != nil {
		tcp = tls.NewListener(tcp, hsv.TLSConfig)
		log.Infof("HTTPs Serving %s ...", hsv.Addr)
	} else {
		log.Infof("HTTP Serving %s ...", hsv.Addr)
	}

	if err := hsv.Serve(tcp); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("HTTP Server %s closed", hsv.Addr)
		} else {
			log.Fatalf(99, "HTTP.Serve(%s) failed: %v", hsv.Addr, err)
		}
	}
}

// Shutdown gracefully shutdown the http servers with timeout '[server] shutdownTimeout' (default 15 seconds).
func (s *Server) Shutdown(wg *sync.WaitGroup) {
	// shutdown http servers
	for _, hsv := range s.HSVs {
		wg.Add(1)
		go s.shutdown(hsv, wg)
	}
}

func (s *Server) shutdown(hsv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()

	// The context is used to inform the server it has some seconds to finish
	// the request it is currently handling
	timeout := ini.GetDuration("server", "shutdownTimeout", 30*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Infof("HTTP Server %s shutting down in %v ...", hsv.Addr, timeout)

	if err := hsv.Shutdown(ctx); err != nil {
		log.Errorf("HTTP Server %s failed to shutdown: %v", hsv.Addr, err)
	}
}

// InitServers initialize TCP listener and HTTP server
func InitServers(hh http.Handler, ids ...string) error {
	if len(ids) == 0 {
		ids = []string{""}
	}
	return initServers(hh, ids...)
}

// initServers initialize TCP listeners and HTTP servers
func initServers(hh http.Handler, ids ...string) error {
	for _, id := range ids {
		srv := &Server{ID: id}
		if err := srv.Init(hh); err != nil {
			return err
		}
		servers[id] = srv
	}
	return nil
}

// ReloadServers reload server configurations
func ReloadServers() error {
	for _, s := range servers {
		s.Config()
	}
	return nil
}

// Serves start serve http servers in go-routines (non-blocking)
func Serves() {
	for _, s := range servers {
		s.Serve()
	}
}

// Shutdowns gracefully shutdown the http servers with timeout '[server] shutdownTimeout' (default 15 seconds).
func Shutdowns() {
	var wg sync.WaitGroup

	for _, s := range servers {
		s.Shutdown(&wg)
	}

	wg.Wait()
}
