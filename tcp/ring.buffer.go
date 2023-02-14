/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import "gopkg.in/hedzr/go-ringbuf.v1/fast"

//// NewRingBuffer will allocate, initialize, and return a ring buffer
//// with the specified size.
//func NewRingBuffer(size int64) *untitled.RingBuffer {
//	if x, err := untitled.NewBuffer(size); err != nil {
//		logrus.WithError(err).Error("new ring-buffer failed")
//		return nil
//	} else {
//		return x
//	}
//}

// newRingBuf will allocate, initialize, and return a ring buffer
// with the specified size.
func newRingBuf(size uint32) fast.RingBuffer {
	if x := fast.New(size); x != nil {
		return x
	}
	return nil
}
