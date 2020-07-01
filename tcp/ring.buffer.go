/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/ringbuf/ringbuf"
	"github.com/sirupsen/logrus"
)

// NewRingBuffer will allocate, initialize, and return a ring buffer
// with the specified size.
func NewRingBuffer(size int64) *ringbuf.RingBuffer {
	if x, err := ringbuf.NewBuffer(size); err != nil {
		logrus.WithError(err).Error("new ring-buffer failed")
		return nil
	} else {
		return x
	}
}
