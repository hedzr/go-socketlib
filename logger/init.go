/*
 * Copyright © 2020 Hedzr Yeh.
 */

package logger

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/logex/formatter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

func silent() bool {
	return cmdr.GetBoolR("quiet")
}

func EarlierInitLogger() {
	l := "OFF"
	if cmdr.InDebugging() {
		l = "DEBUG"
	}
	l = cmdr.GetStringR("mqtool.logger.level", l)

	ll := os.Getenv("APP_LOG")
	if len(ll) > 0 {
		l = ll
		if !silent() {
			Tracef("Using env var APP_LOG level: %+v", ll)
		}
	} else {
		Tracef("Using logging level: %+v", l)
	}

	level := stringToLevel(l)

	// In Goland, you can enable this under 'Run/Debug Configurations', by
	// adding the following into 'Go tool arguments:'
	//
	// -tags=delve
	//
	if cmdr.InDebugging() && level < logrus.DebugLevel {
		level = logrus.DebugLevel
		l = "DEBUG"
	}

	logrus.SetLevel(level)
	if l == "OFF" {
		logrus.SetOutput(ioutil.Discard)
	} else {
		fmtr := cmdr.GetStringR("mqtool.logger.format", "text")

		// ll := os.Getenv("ENT_LOG_FMT")
		// if len(ll) > 0 {
		// 	fmtr = ll
		// }

		switch fmtr {
		case "text":
			logrus.SetFormatter(&formatter.TextFormatter{
				ForceColors:     true,
				DisableColors:   false,
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05.000",
			})
		default:
			logrus.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat:  "2006-01-02 15:04:05.000",
				DisableTimestamp: false,
				PrettyPrint:      false,
			})
		}
	}
}

func stringToLevel(s string) logrus.Level {
	s = strings.ToUpper(s)
	switch s {
	case "OFF", "0", "DISABLE", "DISABLED":
		return logrus.PanicLevel
	case "TRACE":
		return logrus.TraceLevel
	case "DEBUG", "devel", "dev":
		return logrus.DebugLevel
	case "INFO":
		return logrus.InfoLevel
	case "WARN":
		return logrus.WarnLevel
	case "ERROR":
		return logrus.ErrorLevel
	case "FATAL":
		return logrus.FatalLevel
	case "PANIC":
		return logrus.PanicLevel
	default:
		return logrus.FatalLevel
	}
}
