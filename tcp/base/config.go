package base

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
	"gopkg.in/hedzr/errors.v2"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LoggerConfig *log.LoggerConfig
	log.Logger

	Addr                string
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

	prefix := strings.Join([]string{prefixPrefix, s, "tls"}, ".")
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

func (c *Config) BuildLogger() {
	c.Logger = build.New(c.LoggerConfig)
}

func (c *Config) BuildServerAddr() (err error) {
	var host, port string
	if c.Addr == "" {
		c.Addr = cmdr.GetStringRP(c.PrefixInCommandLine, "addr", ":"+cmdr.GetStringRP(c.PrefixInCommandLine, "port", "1024"))
	}
	host, port, err = net.SplitHostPort(c.Addr)
	if port == "" {
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
	host, port, err = net.SplitHostPort(c.Addr)
	if port == "" {
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

func (c *Config) PressEnterToExit() {
	fmt.Print("Press 'Enter' to exit...")
	defer func() { fmt.Println() }()
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}

const DefaultPidPathTemplate = "/var/run/$APPNAME/$APPNAME.pid"
const DefaultPidDirTemplate = "/var/run/$APPNAME"
