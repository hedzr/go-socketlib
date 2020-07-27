package server

import "github.com/hedzr/logex"

type Config struct {
	Addr                string
	LoggerConfig        *logex.LoggerConfig
	PrefixInCommandLine string
	PrefixInConfigFile  string
	PidDir              string
}

func NewConfig() *Config {
	return &Config{
		Addr:                "",
		LoggerConfig:        logex.NewLoggerConfig(),
		PrefixInCommandLine: "tcp.server",
		PrefixInConfigFile:  "tcp.server",
		PidDir:              DefaultPidDirTemplate,
	}
}
