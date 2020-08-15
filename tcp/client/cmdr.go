/*
 * Copyright © 2020 Hedzr Yeh.
 */

package client

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"os"
	"time"
)

const (
	DefaultPort = 8883
)

type CmdrOpt func(*builder)

func WithCmdrUDPMode(mode bool) CmdrOpt {
	return func(b *builder) {
		b.udpMode = mode
	}
}

func WithCmdrPort(port int) CmdrOpt {
	return func(b *builder) {
		b.port = port
	}
}

func WithCmdrPrefixPrefix(prefixPrefix string) CmdrOpt {
	return func(b *builder) {
		b.prefixPrefix = prefixPrefix
	}
}

func WithCmdrClientProtocolInterceptor(fn protocol.ClientInterceptor) CmdrOpt {
	return func(b *builder) {
		picOpt := WithClientProtocolInterceptor(fn)
		b.opts = append(b.opts, picOpt)
	}
}

func WithCmdrClientOptions(opts ...Opt) CmdrOpt {
	return func(b *builder) {
		b.opts = append(b.opts, opts...)
	}
}

func WithCmdrInteractiveCommand(enabled bool) CmdrOpt {
	return func(b *builder) {
		b.interactiveCommand = enabled
	}
}

func WithCmdrCommandAction(action CommandAction) CmdrOpt {
	return func(b *builder) {
		b.action = action
	}
}

func WithCmdrMainLoop(mainLoop MainLoop) CmdrOpt {
	return func(b *builder) {
		b.mainLoop = mainLoop
	}
}

func WithCmdrNil() CmdrOpt {
	return nil
}

type builder struct {
	port               int
	interactiveCommand bool
	action             CommandAction
	mainLoop           MainLoop
	udpMode            bool
	prefixPrefix       string
	opts               []Opt
}

func AttachToCmdr(tcp cmdr.OptCmd, opts ...CmdrOpt) {
	// tcp := root.NewSubCommand().
	// 	Titles("t", "tcp").
	// 	Description("", "").
	// 	Group("Test")
	// // Action(func(cmd *cmdr.Command, args []string) (err error) {
	// // 	return
	// // })

	b := &builder{
		port:   DefaultPort,
		action: runAsCliTool,
		// mainLoop: defaultMainLoop,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}

	if b.interactiveCommand {
		tc2 := tcp.NewSubCommand("interactive-client", "ic").
			Description("TCP interactive client operations").
			Group("Test").
			Action(interactiveRunAsCliTool)
		b.attachTcpClientFlags(tc2)
	}

	if b.mainLoop == nil {
		if b.udpMode {
			b.mainLoop = defaultUdpMainLoop
		} else {
			b.mainLoop = defaultMainLoop
		}
	}

	theClient := tcp.NewSubCommand("client", "c").
		Description("TCP/UDP/Unix client operations").
		// Group("Test").
		Action(func(cmd *cmdr.Command, args []string) (err error) {
			err = b.action(cmd, args, b.mainLoop, b.prefixPrefix, b.opts...)
			return
		})

	b.attachTcpClientFlags(theClient)

	if !b.udpMode {
		b.attachTcpTLSClientFlags(theClient)
	}
}

func defaultMainLoop(ctx context.Context, conn base.Conn, done chan bool, config *base.Config) {
	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		config.Logger.Debugf("signal[%v] caught and exiting this program", s)
	})()
}

func defaultUdpMainLoop(ctx context.Context, conn base.Conn, done chan bool, config *base.Config) {
	var err error

	_, err = conn.RawWrite(ctx, []byte("hello"))
	//uo.WriteTo(nil, []byte("hello"))
	config.Logger.Debugf("'hello' wrote: %v", err)

	if wr, ok := conn.(base.CachedUDPWriter); ok {
		//_, err = uo.WriteThrough([]byte("world"))
		wr.WriteTo(nil, []byte("world"))
		config.Logger.Debugf("'world' wrote: %v", err)
	}

	time.Sleep(time.Second)
	config.PressEnterToExit()
	// _, _ = uo.WriteThrough([]byte("hello"))

	//n, data := 0, make([]byte, 1024)
	//n, err = conn.Read(data)
	//fmt.Printf("read %s from <%s>\n", data[:n], conn.RemoteAddr())
}

func (b *builder) attachTcpClientFlags(theClient cmdr.OptCmd) {

	network := "tcp"
	if b.udpMode {
		// b.opts = append(b.opts, WithServerUDPMode(true))
		network = "udp"
	}

	cmdr.NewInt(b.port).
		Titles("port", "p").
		Description("The port to connect to").
		Group("Test").
		Placeholder("PORT").
		AttachTo(theClient)

	cmdr.NewString("127.0.0.1").
		Titles("host", "h", "address", "addr").
		Description("The hostname or IP to connect to").
		Group("Test").
		Placeholder("HOST-or-IP").
		AttachTo(theClient)
	// don't use localhost, it may cause 'lookup localhost: no such host' error in debug mode.

	cmdr.NewInt(100).
		Titles("times", "t").
		Description("repeat sending times").
		Group("Test").
		Placeholder("n").
		AttachTo(theClient)

	cmdr.NewInt(3).
		Titles("parallel", "r").
		Description("how many clients parallel").
		Group("Test").
		Placeholder("n").
		AttachTo(theClient)

	cmdr.NewDuration(time.Duration(0)).
		Titles("sleep", "").
		Description("sleep time between each sending").
		Group("Test").
		AttachTo(theClient)

	cmdr.NewBool().
		Titles("interactive", "i").
		Description("run client in interactive mode").
		Group("Test").
		AttachTo(theClient)

	cmdr.NewString(network).
		Titles("0007.network", "").
		Description("network: tcp, tcp4, tcp6, unix, unixpacket", `

// The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
//
// For TCP networks, if the host in the address parameter is empty or
// a literal unspecified IP address, Listen listens on all available
// unicast and anycast IP addresses of the local system.
// To only use IPv4, use network "tcp4".
// The address can use a host name, but this is not recommended,
// because it will create a listener for at most one of the host's IP
// addresses.
// If the port in the address parameter is empty or "0", as in
// "127.0.0.1:" or "[::1]:0", a port number is automatically chosen.
// The Addr method of Listener can be used to discover the chosen
// port.

`).
		// Group("TLS").
		AttachTo(theClient)

}

func (b *builder) attachTcpTLSClientFlags(theClient cmdr.OptCmd) {
	theClient.NewFlagV(false, "enable-tls", "tls").
		Description("enable TLS mode").
		Group("TLS")

	cmdr.NewString("root.pem").
		Titles("cacert", "ca").
		Description("CA cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH").
		AttachTo(theClient)
	cmdr.NewString("cert.pem").
		Titles("server-cert", "sc").
		Description("server public-cert path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH").
		AttachTo(theClient)
	cmdr.NewString("client.pem").
		Titles("cert", "cert").
		Description("[client-auth] client public-cert path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH").
		AttachTo(theClient)
	cmdr.NewString("client.key").
		Titles("key", "key").
		Description("[client-auth] client private-key path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH").
		AttachTo(theClient)
	cmdr.NewBool().
		Titles("client-auth", "").
		Description("[client-auth] enable client cert authentication").
		Group("TLS").
		AttachTo(theClient)
	cmdr.NewBool().
		Titles("insecure", "k").
		Description("[client-auth] ignore server cert validation (for self-signed server)").
		Group("TLS").
		AttachTo(theClient)
	cmdr.NewInt(2).
		Titles("tls-version", "").
		Description("tls-version: 0,1,2,3").
		Group("TLS").
		AttachTo(theClient)

}
