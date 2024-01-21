//go:build !windows && !appengine
// +build !windows,!appengine

/*
 * Copyright © 2020 Hedzr Yeh.
 */

package net

import (
	"errors"
	"io"
	"net"
	"syscall"
)

var errUnexpectedRead = errors.New("unexpected read from socket")

// checkConn trying to inspect the underlying system tcp/udp/unix connection (socket) to dig the state of conn
func checkConn(conn net.Conn) error {
	var sysErr error

	sysConn, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}
	rawConn, err := sysConn.SyscallConn()
	if err != nil {
		return err
	}

	err = rawConn.Read(func(fd uintptr) bool {
		var buf [1]byte
		n, err := syscall.Read(int(fd), buf[:])
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
		case n > 0:
			sysErr = errUnexpectedRead
		case errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK):
			sysErr = nil
		default:
			sysErr = err
		}
		return true
	})
	if err != nil {
		return err
	}

	return sysErr
}
