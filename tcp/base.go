/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/socketlib/logger"
)

type base struct{ logger.Base }

func newBase(tag string) base {
	return base{Base: logger.Base{Tag: tag}}
}
