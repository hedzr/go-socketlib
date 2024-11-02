//go:build windows || appengine
// +build windows appengine

/*
 * Copyright © 2020 Hedzr Yeh.
 */

package net

import "net"

func checkConn(conn net.Conn) error {
	return nil
}
