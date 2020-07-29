package base

import (
	"github.com/hedzr/log"
)

type Config struct {
	Addr                string
	Adapter             string // network adapter name. such as "en4". default "". for udp multicast
	LoggerConfig        *log.LoggerConfig
	PrefixInCommandLine string
	PrefixInConfigFile  string
	PidDir              string
}

func NewConfig() *Config {
	return &Config{
		Addr:                "",
		Adapter:             "",
		LoggerConfig:        log.NewLoggerConfig(),
		PrefixInCommandLine: "tcp.server",
		PrefixInConfigFile:  "tcp.server",
		PidDir:              DefaultPidDirTemplate,
	}
}

const DefaultPidPathTemplate = "/var/run/$APPNAME/$APPNAME.pid"
const DefaultPidDirTemplate = "/var/run/$APPNAME"
