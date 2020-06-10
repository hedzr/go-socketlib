/*
 * Copyright © 2020 Hedzr Yeh.
 */

package internal

import (
	"log"
	"net"
	"os"
	"syscall"
)

func IsCaredNetError(err error) bool {
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}

	if netErr.Timeout() {
		log.Println("timeout")
		return true
	}

	opErr, ok := netErr.(*net.OpError)
	if !ok {
		return false
	}

	switch t := opErr.Err.(type) {
	case *net.DNSError:
		log.Printf("net.DNSError:%+v", t)
		return true
	case *os.SyscallError:
		log.Printf("os.SyscallError:%+v", t)
		if errno, ok := t.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ECONNREFUSED:
				log.Println("connect refused")
				return true
			case syscall.ETIMEDOUT:
				log.Println("timeout")
				return true
			}
		}
	default:
		log.Printf("err type = %+v; err = %+v", t, err)
	}

	return false
}
