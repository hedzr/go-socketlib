/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"io/ioutil"
	"net"
	"path"
	"strings"
	"time"
)

func NewTlsConfig(fn func(config *CmdrTlsConfig)) *CmdrTlsConfig {
	s := &CmdrTlsConfig{
		DialTimeout: 10 * time.Second,
	}
	if fn != nil {
		fn(s)
	}
	return s
}

func NewCmdrTlsConfig(prefixInConfigFile, prefixInCommandline string) *CmdrTlsConfig {
	s := &CmdrTlsConfig{
		DialTimeout: 10 * time.Second,
	}
	if len(prefixInConfigFile) > 0 {
		s.InitTlsConfigFromConfigFile(prefixInConfigFile)
	}
	if len(prefixInCommandline) > 0 {
		s.InitTlsConfigFromCommandline(prefixInCommandline)
	}
	return s
}

func (s *CmdrTlsConfig) WithLogger(logger log.Logger) *CmdrTlsConfig {
	var t CmdrTlsConfig
	ptr := cmdr.Clone(s, &t)
	p := ptr.(*CmdrTlsConfig)
	p.logger = logger
	return p
}

func (s *CmdrTlsConfig) String() string {
	var sb strings.Builder
	sb.WriteString("[CmdrTlsConfig: ")
	sb.WriteString(fmt.Sprintf("Enabled: %v", s.Enabled))
	if s.Enabled {
		sb.WriteString(fmt.Sprintf(", CA cert: %v", s.CaCert))
		sb.WriteString(fmt.Sprintf(", Server cert (client-side): %v", s.ServerCert))
		sb.WriteString(fmt.Sprintf(", Server cert bundle/Client cert: %v", s.Cert))
		sb.WriteString(fmt.Sprintf(", Server/Client key: %v", s.Key))
		sb.WriteString(fmt.Sprintf(", Client Auth: %v", s.ClientAuth))
		sb.WriteString(fmt.Sprintf(", Min TLS: %v", s.MinTlsVersion))
	}
	sb.WriteString("]")
	return sb.String()
}

func (s *CmdrTlsConfig) IsServerCertValid() bool {
	return s.ServerCert != "" || s.CaCert != ""
}

func (s *CmdrTlsConfig) IsCertValid() bool {
	return s.Cert != "" && s.Key != ""
}

func (s *CmdrTlsConfig) IsClientAuthValid() bool {
	return s.ClientAuth && s.Cert != "" && s.Key != ""
}

func (s *CmdrTlsConfig) InitTlsConfigFromCommandline(prefix string) {
	var b bool
	var sz string

	s.Enabled = cmdr.GetBoolRP(prefix, "enable-tls")
	if !s.Enabled {
		return
	}

	b = cmdr.GetBoolRP(prefix, "client-auth")
	if b {
		s.ClientAuth = b
	}
	sz = cmdr.GetStringRP(prefix, "cacert")
	if sz != "" {
		s.CaCert = sz
	}
	sz = cmdr.GetStringRP(prefix, "cert")
	if sz != "" {
		s.Cert = sz
	}
	sz = cmdr.GetStringRP(prefix, "server-cert")
	if sz != "" {
		s.ServerCert = sz
	}
	b = cmdr.GetBoolRP(prefix, "insecure")
	if b {
		s.InsecureSkipVerify = b
	}
	sz = cmdr.GetStringRP(prefix, "key")
	if sz != "" {
		s.Key = sz
	}

	for _, loc := range cmdr.GetStringSliceRP(prefix, "locations", "./ci/certs", "$CFG_DIR/certs") {
		if s.CaCert != "" && cmdr.FileExists(path.Join(loc, s.CaCert)) {
			s.CaCert = path.Join(loc, s.CaCert)
		} else if s.CaCert != "" {
			continue
		}
		if s.ServerCert != "" && cmdr.FileExists(path.Join(loc, s.ServerCert)) {
			s.ServerCert = path.Join(loc, s.ServerCert)
		}
		if s.Cert != "" && cmdr.FileExists(path.Join(loc, s.Cert)) {
			s.Cert = path.Join(loc, s.Cert)
		} else if s.Cert != "" {
			continue
		}
		if s.Key != "" && cmdr.FileExists(path.Join(loc, s.Key)) {
			s.Key = path.Join(loc, s.Key)
		} else if s.Key != "" {
			continue
		}
	}

	switch cmdr.GetIntRP(prefix, "tls-version", 2) {
	case 0:
		s.MinTlsVersion = VersionTLS10
	case 1:
		s.MinTlsVersion = VersionTLS11
	case 3:
		s.MinTlsVersion = VersionTLS13
	default:
		s.MinTlsVersion = VersionTLS12
	}
}

func (s *CmdrTlsConfig) InitTlsConfigFromConfigFile(prefix string) {
	// prefix := "mqtt.server.tls"
	// tls:
	//   enabled: true
	//   cacert: root.pem
	//   cert: cert.pem
	//   key: cert.key
	//   locations:
	// 	   - ./ci/certs
	// 	   - $CFG_DIR/certs
	s.Enabled = cmdr.GetBoolRP(prefix, "enabled")
	if s.Enabled {
		s.ClientAuth = cmdr.GetBoolRP(prefix, "client-auth")
		s.CaCert = cmdr.GetStringRP(prefix, "cacert")
		s.Cert = cmdr.GetStringRP(prefix, "cert")
		s.Key = cmdr.GetStringRP(prefix, "key")

		for _, loc := range cmdr.GetStringSliceRP(prefix, "locations") {
			if s.CaCert != "" && cmdr.FileExists(path.Join(loc, s.CaCert)) {
				s.CaCert = path.Join(loc, s.CaCert)
			} else if s.CaCert != "" {
				continue
			}
			if s.Cert != "" && cmdr.FileExists(path.Join(loc, s.Cert)) {
				s.Cert = path.Join(loc, s.Cert)
			} else if s.Cert != "" {
				continue
			}
			if s.Key != "" && cmdr.FileExists(path.Join(loc, s.Key)) {
				s.Key = path.Join(loc, s.Key)
			} else if s.Key != "" {
				continue
			}
		}

		switch cmdr.GetIntRP(prefix, "tls-version", int(s.MinTlsVersion-tls.VersionTLS10)) {
		case 0:
			s.MinTlsVersion = VersionTLS10
		case 1:
			s.MinTlsVersion = VersionTLS11
		case 3:
			s.MinTlsVersion = VersionTLS13
		default:
			s.MinTlsVersion = VersionTLS12
		}
	}
}

// ToServerTlsConfig builds an tls.Config object for server.Serve
func (s *CmdrTlsConfig) ToServerTlsConfig() (config *tls.Config) {
	var err error
	config, err = s.newTlsConfig()
	if err == nil {
		if s.CaCert != "" {
			var rootPEM []byte
			rootPEM, err = ioutil.ReadFile(s.CaCert)
			if err != nil || rootPEM == nil {
				return
			}
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(rootPEM)
			if ok {
				config.ClientCAs = pool
			}
		}
	}
	return config
}

func (s *CmdrTlsConfig) ToTlsConfig() (config *tls.Config) {
	config, _ = s.newTlsConfig()
	return config
}

func (s *CmdrTlsConfig) newTlsConfig() (config *tls.Config, err error) {
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(s.Cert, s.Key)
	if err != nil {
		err = errors.New("error parsing X509 certificate/key pair").Attach(err)
		return
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		err = errors.New("error parsing certificate").Attach(err)
		return
	}

	// Create TLSConfig
	// We will determine the cipher suites that we prefer.
	config = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   uint16(s.MinTlsVersion),
	}

	// Require client certificates as needed
	if s.IsClientAuthValid() {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// Add in CAs if applicable.
	if s.ClientAuth {
		if s.CaCert != "" {
			var rootPEM []byte
			rootPEM, err = ioutil.ReadFile(s.CaCert)
			if err != nil || rootPEM == nil {
				return nil, err
			}
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(rootPEM)
			if !ok {
				err = errors.New("failed to parse root ca certificate")
			}
			config.ClientCAs = pool
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if err != nil {
		config = nil
	}
	return
}

func (s *CmdrTlsConfig) NewTlsListener(l net.Listener) (listener net.Listener, err error) {
	if s != nil && s.IsCertValid() {
		var config *tls.Config
		config, err = s.newTlsConfig()
		if err != nil {
			if s.logger != nil {
				s.logger.Fatalf("fatal error: %v", err)
			}
			return
		}
		listener = tls.NewListener(l, config)
	}
	return
}

// Dial connects to the given network address using net.Dial
// and then initiates a TLS handshake, returning the resulting
// TLS connection.
// Dial interprets a nil configuration as equivalent to
// the zero configuration; see the documentation of Config
// for the defaults.
func (s *CmdrTlsConfig) Dial(network, addr string) (conn net.Conn, err error) {
	if s != nil && s.IsServerCertValid() {
		roots := x509.NewCertPool()

		err = s.addCert(roots, s.ServerCert)
		if err != nil {
			return
		}
		err = s.addCert(roots, s.CaCert)
		if err != nil {
			return
		}

		cfg := &tls.Config{
			RootCAs: roots,
		}

		if s.IsClientAuthValid() {
			var cert tls.Certificate
			cert, err = tls.LoadX509KeyPair(s.Cert, s.Key)
			if err != nil {
				return
			}
			cfg.Certificates = []tls.Certificate{cert}
		}

		cfg.InsecureSkipVerify = s.InsecureSkipVerify

		if s.logger != nil {
			s.logger.Printf("Connecting to %s over TLS [-k=%v]...\n", addr, cfg.InsecureSkipVerify)
		}

		dialer := &net.Dialer{Timeout: s.DialTimeout}
		// Use the tls.Config here in http.Transport.TLSClientConfig
		conn, err = tls.DialWithDialer(dialer, network, addr, cfg)
	} else {
		if s.logger != nil {
			s.logger.Printf("Connecting to %s...\n", addr)
		}
		conn, err = net.DialTimeout(network, addr, s.DialTimeout)
	}
	return
}

func (s *CmdrTlsConfig) addCert(roots *x509.CertPool, certPath string) (err error) {
	if certPath != "" {
		var rootPEM []byte
		rootPEM, err = ioutil.ReadFile(certPath)
		if err != nil {
			return
		}

		ok := roots.AppendCertsFromPEM(rootPEM)
		if !ok {
			// panic("failed to parse root certificate")
			err = errors.New("failed to parse root certificate")
			return
		}
	}
	return
}
