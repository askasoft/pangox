package xcert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	// TLSConfig
	TLSConfig = &tls.Config{
		GetCertificate: GetCertificate,
		NextProtos: []string{
			"h2", "http/1.1", // enable HTTP/2
			acme.ALPNProto, // enable tls-alpn ACME challenges
		},
	}

	// Certificate X509 KeyPair
	Certificate *tls.Certificate

	// Certificate Manager
	CertManager *autocert.Manager
)

func GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if Certificate != nil {
		return Certificate, nil
	}

	if CertManager != nil {
		return CertManager.GetCertificate(chi)
	}

	return nil, errors.New("nil certificate")
}

func InitCertificate() {
	certificate := ini.GetString("server", "certificate")

	switch certificate {
	case "":
	case "autocert":
		certCache := ini.GetString("server", "autocertcache", "./cert")
		certHosts := str.Fields(ini.GetString("server", "autocerthosts"))
		CertManager = MakeCertManager(certCache, certHosts...)
	default:
		certkeyfile := ini.GetString("server", "certkeyfile")
		xcert, err := LoadCertificate(certificate, certkeyfile)
		if err != nil {
			log.Fatal(98, err)
		}
		Certificate = xcert
	}
}

func ReloadCertificate() {
	certificate := ini.GetString("server", "certificate")

	switch certificate {
	case "":
		Certificate = nil
		CertManager = nil
	case "autocert":
		certCache := ini.GetString("server", "autocertcache", "./cert")
		certHosts := str.Fields(ini.GetString("server", "autocerthosts"))
		if CertManager == nil {
			CertManager = MakeCertManager(certCache, certHosts...)
		} else {
			CertManager.Cache = autocert.DirCache(certCache)
			CertManager.HostPolicy = MakeHostPolicy(certHosts...)
		}
		Certificate = nil
	default:
		certkeyfile := ini.GetString("server", "certkeyfile")
		xcert, err := LoadCertificate(certificate, certkeyfile)
		if err != nil {
			log.Error(err)
			return
		}
		Certificate = xcert
	}
}

func AcceptAnyHost(ctx context.Context, host string) error {
	return nil
}

func MakeHostPolicy(hosts ...string) autocert.HostPolicy {
	if len(hosts) > 0 {
		return autocert.HostWhitelist(hosts...)
	}
	return AcceptAnyHost
}

func MakeCertManager(cachedir string, hosts ...string) *autocert.Manager {
	acc := autocert.DirCache(cachedir)

	acm := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      acc,
		HostPolicy: MakeHostPolicy(hosts...),
	}
	return acm
}

func LoadCertificate(certificate, certkeyfile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certificate, certkeyfile)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate (%q, %q): %w", certificate, certkeyfile, err)
	}

	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("invalid certificate (%q, %q): %w", certificate, certkeyfile, err)
	}

	return &cert, nil
}
