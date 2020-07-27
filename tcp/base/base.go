package base

import (
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/build"
)

type base struct {
	logex.Logger
}

func newBase(config *logex.LoggerConfig) base {
	return base{
		build.New(config),
	}
}
