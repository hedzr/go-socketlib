package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
)

func newClientObj(conn net.Conn, tid int, logger log.Logger, opts ...Opt) (c *clientObj) {
	c = &clientObj{
		Logger:             logger,
		tid:                tid,
		quiting:            false,
		prefixInConfigFile: "tcp.client.tls",
	}
	if conn != nil {
		c.baseConn = &connWrapper{nil, conn, logger}
	}
	for _, opt := range opts {
		opt(c)
	}

	return
}

type clientObj struct {
	log.Logger
	tid                 int                  // thread-id in parallel testing, 0 for udp, 1..n for tcp
	baseConn            base.Conn            //
	protocolInterceptor protocol.Interceptor //
	quiting             bool
	closeErr            error
	prefixInConfigFile  string
	mainLoop            MainLoop
	buildPackage        BuildPackageFunc
}

func (c *clientObj) ProtocolInterceptor() protocol.Interceptor {
	return c.protocolInterceptor
}

func (c *clientObj) SetProtocolInterceptor(pi protocol.Interceptor) {
	c.protocolInterceptor = pi
}

func (c *clientObj) SetBaseConn(bc base.Conn) {
	c.baseConn = bc
}

func (c *clientObj) AsBaseConn() base.Conn {
	return c.baseConn
}

func (c *clientObj) Join(ctx context.Context, done chan<- bool) {
	if c.baseConn != nil {
		c.Close()
		close(done)
	}
}

func (c *clientObj) Close() {
	c.quiting = true
	if cc := c.baseConn; cc != nil {
		c.baseConn = nil
		if c.protocolInterceptor != nil {
			c.protocolInterceptor.OnClosing(cc, 0)
		}
		c.Debugf("closing c.baseConn")
		cc.Close()
		if c.protocolInterceptor != nil {
			c.protocolInterceptor.OnClosed(cc, 0)
		}
	}
}

func (c *clientObj) startLoopers(done chan bool, readBroken chan<- bool) {
	go c.readConnection(done, readBroken)
	go c.runPrompt()
}

func (c *clientObj) run(done chan bool, readBroken chan<- bool) {
	go c.readConnection(done, readBroken)
	c.runPrompt()
}

func (c *clientObj) runPrompt() {
	fmt.Println("type 'quit' to exit client, '/quit' to exit both server and client.")
	defer c.Close()
	for c.quiting == false {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && c.quiting {
				break
			}
			c.Errorf("TCP i/o read failed: %v", err)
		}

		if text == "quit\n" || text == "exit\n" {
			c.quiting = true
			break
		}

		if text == "/quit\n" {
			c.quiting = true
		}

		_, err = c.baseConn.WriteNow([]byte(text), time.Second)
		if err != nil {
			c.Errorf("error writing to stream: %v", err)
			break
		}
	}
}

func (c *clientObj) readConnection(done chan bool, readBroken chan<- bool) {
	defer func() {
		readBroken <- true
	}()

	for {
		scanner := bufio.NewScanner(c.baseConn)

		for c.quiting == false {
			ok := scanner.Scan()

			select {
			case <-done:
				return
			default:
			}

			text := scanner.Text()

			command := handleCommands(text)
			if !command {
				fmt.Printf("\b\b** %s\n> ", text)
			}

			if !ok {
				if c.quiting {
					return
				}
				// fmt.Println("Reached EOF on server connection.")
				c.Errorf("[%d] <%v - %v> %v", c.tid, c.baseConn.LocalAddr(), c.baseConn.RemoteAddr(), "Reached EOF on server connection.")
				return
			}
		}
	}
}

func handleCommands(text string) bool {
	r, err := regexp.Compile("^%.*%$")
	if err == nil {
		if r.MatchString(text) {

			switch {
			case text == "%quit%":
				fmt.Println("\b\bServer is leaving. Hanging up.")
				os.Exit(0)
			}

			return true
		}
	}

	return false
}
