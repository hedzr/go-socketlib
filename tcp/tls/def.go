/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tls

import (
	"github.com/hedzr/log"
	"time"
)

// CmdrTlsConfig wraps the certificates.
// For server-side, the `Cert` field must be a bundle of server certificates with all root CAs chain.
// For server-side, the `CaCert` is optional for extra client CA's.
type CmdrTlsConfig struct {
	Enabled            bool          // Both
	CaCert             string        // server-side: optional server's CA;   client-side: client's CA
	ServerCert         string        //                                      client-side: the server's cert
	Cert               string        // server-side: server's cert bundle;   client-side: client's cert
	Key                string        // server-side: server's key;           client-side: client's key
	ClientAuth         bool          // Both
	InsecureSkipVerify bool          // client-side only
	MinTlsVersion      VersionTLS    // Both
	DialTimeout        time.Duration // for dialing

	logger log.Logger
}

type Initializer func(config *CmdrTlsConfig)
