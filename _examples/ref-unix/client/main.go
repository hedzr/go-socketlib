package main

import (
	"net"
	"time"
)

func main() {
	c, err := net.Dial("unix", "/tmp/echo.sock")
	if err != nil {
		panic(err.Error())
	}
	for {
		_, err = c.Write([]byte("hi\n"))
		if err != nil {
			println(err.Error())
		}
		time.Sleep(1e9)
	}
}
