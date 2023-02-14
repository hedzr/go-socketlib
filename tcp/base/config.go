package base

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/hedzr/cmdr"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
	"gopkg.in/hedzr/errors.v3"
)

type Config struct {
	LoggerConfig *log.LoggerConfig
	log.Logger

	Addr                string
	Uri                 *url.URL
	UriBase             string
	Adapter             string // network adapter name. such as "en4". default "". for udp multicast
	PrefixInCommandLine string
	PrefixInConfigFile  string
	PidDir              string

	Network              string
	TlsConfigInitializer tls2.Initializer
}

const defaultNetType = "tcp"

func NewConfigFromCmdrCommand(isServer bool, prefixPrefix string, cmd *cmdr.Command) *Config {
	prefixCLI := cmd.GetDottedNamePath()

	// tcp, udp, unix, ...
	netType := cmdr.GetStringRP(prefixCLI, "network", defaultNetType)
	//Network = cmdr.GetStringRP(Network + "." + prefixSuffix, "network", Network)

	loggerConfig := cmdr.NewLoggerConfig()

	return NewConfigWithParams(isServer, netType,
		prefixPrefix, prefixCLI,
		loggerConfig,
		nil,
		"", "", "")
}

func NewConfigWithParams(isServer bool, netType, prefixPrefix, prefixCLI string, loggerConfig *log.LoggerConfig, tlsConfigInitializer tls2.Initializer, addr, uriBase, adapter string) *Config {
	s := "client"
	if isServer {
		s = "server"
	}

	if prefixPrefix == "" {
		prefixPrefix = "tcp"
	}

	prefix := strings.Join([]string{prefixPrefix, s}, ".")
	if prefixCLI == "" {
		prefixCLI = strings.Join([]string{prefixPrefix, s}, ".")
	}

	return &Config{
		Addr:                 addr,
		Adapter:              adapter,
		UriBase:              uriBase,
		LoggerConfig:         loggerConfig,
		PrefixInCommandLine:  prefixCLI,
		PrefixInConfigFile:   prefix,
		PidDir:               DefaultPidDirTemplate,
		Network:              netType,
		TlsConfigInitializer: tlsConfigInitializer,
	}
}

func (c *Config) UpdatePrefixInConfigFile(s string) {
	if len(s) > 0 {
		c.PrefixInConfigFile = s
	}
}

func (c *Config) BuildLogger() {
	c.Logger = build.New(c.LoggerConfig)
}

func (c *Config) BuildPidFileStruct() *pidFileStruct {
	return makePidFS(c.PrefixInCommandLine, c.PrefixInConfigFile, c.PidDir)
}

func (c *Config) BuildServerAddr() (err error) {
	var host, port string
	if c.Addr == "" {
		c.Addr = cmdr.GetStringRP(c.PrefixInCommandLine, "host", ":"+cmdr.GetStringRP(c.PrefixInCommandLine, "port", "1024"))
	}
	host, port, err = net.SplitHostPort(c.Addr)
	if port == "" || port == "0" {
		err = nil
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInConfigFile, "ports.default"), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInCommandLine, "port", 1024), 10)
		if port == "0" {
			err = errors.New("invalid port number: %q", port)
			return
		}
	}
	//if host == "" {
	//	host = "0.0.0.0"
	//	// forceIPv6 make all IPv6 ip-addresses of this PC are listened, instead of its IPv4 addresses
	//	const forceIPv6 = false
	//	if forceIPv6 {
	//		host = "[::]"
	//	}
	//}
	c.Addr = net.JoinHostPort(host, port)
	c.UriBase = cmdr.GetStringRP(c.PrefixInConfigFile, "base-uri")
	return
}

func (c *Config) BuildAddr() (err error) {
	var host, port string
	if c.Addr == "" {
		c.Addr = cmdr.GetStringRP(c.PrefixInCommandLine, "host", ":"+cmdr.GetStringRP(c.PrefixInCommandLine, "port", "1024"))
	}
	if strings.Contains(c.Addr, "://") {
		return c.BuildUriAddr(cmdr.GetStringRP(c.PrefixInCommandLine, "port"))
	}
	c.UriBase = c.Addr

	host = c.Addr //ip := net.ParseIP(host)
	if !strings.Contains(host, ":") {
		host = net.JoinHostPort(host, "0")
	}
	host, port, err = net.SplitHostPort(host)
	if port == "" || port == "0" {
		err = nil
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInCommandLine, "port", 1024), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInConfigFile, "ports.default"), 10)
		if port == "0" {
			err = errors.New("invalid port number: %q", port)
			return
		}
	}
	if host == "" {
		host = "127.0.0.1"
	}
	c.Addr = net.JoinHostPort(host, port)
	return
}

func (c *Config) BuildUriAddr(defaultPort string) (err error) {
	c.Uri, err = url.Parse(cmdr.GetStringRP(c.PrefixInCommandLine, "host", c.Addr))
	if err == nil {
		c.UriBase = c.Uri.String()
		port := c.Uri.Port()
		if port == "" || port == "0" {
			if defaultPort == "" || defaultPort == "0" {
				defaultPort = cmdr.GetStringRP(c.PrefixInConfigFile, "ports.default")
				if strings.HasSuffix(c.Uri.Scheme, "s") {
					defaultPort = cmdr.GetStringRP(c.PrefixInConfigFile, "ports.tls",
						cmdr.GetStringRP(c.PrefixInConfigFile, "ports.dtls"),
					)
				}
			}
			port = defaultPort
		}
		c.Uri.Host = net.JoinHostPort(c.Uri.Hostname(), port)
		c.Addr = c.Uri.Host
	}
	return
}

func (c *Config) PressEnterToExit() {
	fmt.Print("Press 'Enter' to exit...")
	defer func() { fmt.Println() }()
	b := make([]byte, 1)
	_, err := os.Stdin.Read(b)
	if err == nil {
		fmt.Printf("[%v]", string(b))
	}
}

const DefaultPidPathTemplate = "/var/run/$APPNAME/$APPNAME.pid"
const DefaultPidDirTemplate = "/var/run/$APPNAME"
