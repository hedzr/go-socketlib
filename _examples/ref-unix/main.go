package main

import (
	"fmt"
	"net"
	"os"

	logz "log/slog"
)

func echoServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			logz.Error("Read: ", "err", err)
			return
		}

		data := buf[0:nr]
		fmt.Printf("Received: %v", string(data))
		_, err = c.Write(data)
		if err != nil {
			logz.Error("Write: ", "err", err)
			break
		}
	}
}

const sockfile = "/tmp/echo.sock"

type sockfileS struct {
	file string
}

func (s *sockfileS) Close() {
	err := os.Remove(s.file)
	if err != nil {
		logz.Error("delete file error", "err", err, "sockfile", s.file)
		return
	}
	logz.Info("sockfile deleted", "sockfile", s.file)
}

func (s *sockfileS) Listen() (l net.Listener, err error) {
	l, err = net.Listen("unix", s.file)
	return
}

func main() {
	var ss = &sockfileS{sockfile}

	ss.Close() // force deleting last remained sock file when unexpected terminated.

	l, err := ss.Listen()
	if err != nil {
		logz.Error("listen error", "err", err)
		return
	}

	defer l.Close()
	defer ss.Close()

	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			logz.Error("accept error", "err", err)
			return
		}

		go echoServer(conn)
	}
}

func init() {
	// println("OK")
	// logz.SetLevel(logz.DebugLevel)
	// logz.AddFlags(logz.Lprivacypath | logz.Lprivacypathregexp)
}
