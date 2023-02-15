/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/log"
)

// import (
//	"github.com/hedzr/go-socketlib/logger"
// )
//
// type base struct{ logger.Base }
//
// func newBase(tag string) base {
//	return base{Base: logger.Base{Tag: tag}}
// }

type base struct {
	log.Logger
}

func newBaseWithLogger(l log.Logger) base {
	return base{
		l,
	}
}

func newBase(config *log.LoggerConfig) base {
	return base{
		log.GetLogger(), // build.New(config),
	}
}
