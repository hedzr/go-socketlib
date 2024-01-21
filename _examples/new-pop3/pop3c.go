package main

import (
	"context"
	"fmt"
	stdnet "net"
	"time"

	"github.com/hedzr/go-socketlib/net"
	"github.com/hedzr/go-socketlib/net/api"
)

func runClient(ctx context.Context, _ net.Server, _ stdnet.Listener, logger net.Logger) {
	go func() {
		s := newPop3Client(pop3serverAddress, logger)

		if err := s.Connect(ctx); err != nil {
			s.Fatal("cannot connect server", "server.addr", pop3serverAddress, "err", err)
		}

		s.Println("connected.")
		defer s.Close()
		time.Sleep(2 * time.Second)

		s.SendCmd("user user")
		// time.Sleep(1 * time.Second)
		s.SendCmd("pass pass")
		time.Sleep(1 * time.Second)

		s.SendCmd("stat")
		// time.Sleep(2 * time.Second)
		s.SendCmd("list 2")
		time.Sleep(3 * time.Second)

		s.SendCmd("quit")
		time.Sleep(3 * time.Second)
	}()
}

func runDemoClient(ctx context.Context, _ net.Server, _ stdnet.Listener, logger net.Logger) {
	c := net.NewClient()

	if err := c.Dial("tcp", pop3serverAddress); err != nil {
		c.Fatal("connecting to server failed", "err", err, "server-endpoint", ":7099")
	}
	c.Info("[client] connected", "server.addr", c.RemoteAddr())
	c.RunDemo(ctx)
}

type pop3C struct {
	addr string
	net.Client

	net.ExtraPrinters // import ExtraPrinters
	net.Logger        // import Logger
}

func newPop3Client(addr string, logger net.Logger) *pop3C {
	s := &pop3C{addr: addr}
	s.Client = net.NewClient(net.WithClientInterceptor(s), net.WithClientLogger(logger))
	s.Logger = s.Client.(net.Logger)
	s.ExtraPrinters = s.Client.(net.ExtraPrinters)
	return s
}

func (c *pop3C) Connect(ctx context.Context) error {
	if err := c.Dial("tcp", c.addr); err != nil {
		return err
	}
	c.Run(ctx)
	return nil
}

func (c *pop3C) SendCmd(cmd string) {
	data := fmt.Sprintf("%s\r\n", cmd)
	_, _ = c.Write([]byte(data))
}

func (c *pop3C) OnReading(ctx context.Context, conn api.Conn, data []byte, ch chan<- []byte) (processed bool, err error) {
	c.Log(ctx, net.LevelHint, "[pop3C]   OnReading", "how-many-bytes", len(data), "msg", string(data))
	processed = true
	return
}

func (c *pop3C) OnWriting(ctx context.Context, conn api.Conn, data []byte) (processed bool, err error) {
	c.Log(ctx, net.LevelHint, "[pop3C]   OnWriting", "how-many-bytes", len(data), "msg", string(data))
	return
}
