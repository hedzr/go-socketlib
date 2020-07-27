/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/build"
)

//import (
//	"github.com/hedzr/go-socketlib/logger"
//)
//
//type base struct{ logger.Base }
//
//func newBase(tag string) base {
//	return base{Base: logger.Base{Tag: tag}}
//}

type base struct {
	logex.Logger
}

func newBaseWithLogger(l logex.Logger) base {
	return base{
		l,
	}
}

func newBase(config *logex.LoggerConfig) base {
	return base{
		build.New(config),
	}
}
