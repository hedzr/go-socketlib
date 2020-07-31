package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
	"io"
	"net"
	"os"
	"regexp"
	"time"
)

func newClientObj(conn net.Conn, logger log.Logger, opts ...Opt) (c *clientObj) {
	c = &clientObj{
		Logger:             logger,
		conn:               conn,
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
	conn                net.Conn // for tcp, unix
	baseConn            base.Conn
	protocolInterceptor protocol.Interceptor
	quiting             bool
	closeErr            error
	prefixInConfigFile  string
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

func (c *clientObj) Join(ctx context.Context, done chan bool) {
	if c.baseConn != nil {
		c.Close()
		close(done)
	}
}

func (c *clientObj) Close() {
	if c.baseConn != nil {
		if c.protocolInterceptor != nil {
			c.protocolInterceptor.OnClosing(c.baseConn, 0)
		}
		c.baseConn.Close()
		c.conn = nil
		c.baseConn = nil
		if c.protocolInterceptor != nil {
			c.protocolInterceptor.OnClosed(c.baseConn, 0)
		}
	}
}

func (c *clientObj) startLoopers() {
	go c.readConnection()
	go c.runPrompt()
}

func (c *clientObj) run() {
	go c.readConnection()
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
			break
		}

		if text == "/quit\n" {
			c.quiting = true
		}

		err = c.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			c.Errorf("error set writing deadline: %v", err)
			break
		}
		_, err = c.conn.Write([]byte(text))
		if err != nil {
			c.Errorf("error writing to stream: %v", err)
			break
		}
	}
}

func (c *clientObj) readConnection() {
	for {
		scanner := bufio.NewScanner(c.conn)

		for c.quiting == false {
			ok := scanner.Scan()
			text := scanner.Text()

			command := handleCommands(text)
			if !command {
				fmt.Printf("\b\b** %s\n> ", text)
			}

			if !ok {
				if c.quiting {
					return
				}
				fmt.Println("Reached EOF on server connection.")
				break
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
