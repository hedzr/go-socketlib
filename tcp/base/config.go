package base

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
	"gopkg.in/hedzr/errors.v2"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LoggerConfig *log.LoggerConfig
	log.Logger

	Addr                string
	Uri                 *url.URL
	Adapter             string // network adapter name. such as "en4". default "". for udp multicast
	PrefixInCommandLine string
	PrefixInConfigFile  string
	PidDir              string

	Network string
}

const defaultNetType = "tcp"

func NewConfigFromCmdrCommand(isServer bool, prefixPrefix string, cmd *cmdr.Command) *Config {
	prefixCLI := cmd.GetDottedNamePath()

	// tcp, udp, unix, ...
	netType := cmdr.GetStringRP(prefixCLI, "network", defaultNetType)
	//Network = cmdr.GetStringRP(Network + "." + prefixSuffix, "network", Network)

	loggerConfig := cmdr.NewLoggerConfig()

	return NewConfigWithParams(true, netType, prefixPrefix, prefixCLI, loggerConfig)
}

func NewConfigWithParams(isServer bool, netType, prefixPrefix, prefixCLI string, loggerConfig *log.LoggerConfig) *Config {
	s := "client"
	if isServer {
		s = "server"
	}

	prefix := strings.Join([]string{prefixPrefix, s}, ".")
	if prefixCLI == "" {
		prefixCLI = strings.Join([]string{prefixPrefix, s}, ".")
	}

	return &Config{
		Addr:                "",
		Adapter:             "",
		LoggerConfig:        loggerConfig,
		PrefixInCommandLine: prefixCLI,
		PrefixInConfigFile:  prefix,
		PidDir:              DefaultPidDirTemplate,
		Network:             netType,
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

func (c *Config) BuildPidFile() *pidFileStruct {
	return makePidFS(c.PrefixInCommandLine, c.PrefixInConfigFile, c.PidDir)
}

func (c *Config) BuildServerAddr() (err error) {
	var host, port string
	if c.Addr == "" {
		c.Addr = cmdr.GetStringRP(c.PrefixInCommandLine, "host", ":"+cmdr.GetStringRP(c.PrefixInCommandLine, "port", "1024"))
	}
	host, port, err = net.SplitHostPort(c.Addr)
	if port == "" || port == "0" {
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
	host, port, err = net.SplitHostPort(c.Addr)
	if port == "" || port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInConfigFile, "ports.default"), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(c.PrefixInCommandLine, "port", 1024), 10)
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
	c.Uri, err = url.Parse(cmdr.GetStringRP(c.PrefixInCommandLine, "host"))
	if err == nil {
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
		c.Uri.Host = net.JoinHostPort(c.Uri.Host, defaultPort)
		c.Addr = c.Uri.Host
	}
	return
}

func (c *Config) PressEnterToExit() {
	fmt.Print("Press 'Enter' to exit...")
	defer func() { fmt.Println() }()
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}

const DefaultPidPathTemplate = "/var/run/$APPNAME/$APPNAME.pid"
const DefaultPidDirTemplate = "/var/run/$APPNAME"
