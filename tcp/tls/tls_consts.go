package tls

type VersionTLS uint16

const (
	VersionTLS10 VersionTLS = 0x0301
	VersionTLS11 VersionTLS = 0x0302
	VersionTLS12 VersionTLS = 0x0303
	VersionTLS13 VersionTLS = 0x0304

	// Deprecated: SSLv3 is cryptographically broken, and is no longer
	// supported by this package. See golang.org/issue/32716.
	VersionSSL30 VersionTLS = 0x0300
)
