package base

import (
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
)

type Base struct {
	log.Logger
}

func NewBase( /*config *log.LoggerConfig*/ ) *Base {
	return &Base{log.GetLogger()}
}

func NewBaseLogger(config *log.LoggerConfig) *Base {
	return &Base{build.New(config)}
}
