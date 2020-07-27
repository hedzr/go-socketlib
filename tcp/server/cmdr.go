/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"github.com/hedzr/cmdr"
)

type CmdrOpt func(*builder)

type builder struct {
	port int
	opts []Opt
}

func WithPort(port int) CmdrOpt {
	return func(b *builder) {
		b.port = port
	}
}

func WithServerOptions(opts ...Opt) CmdrOpt {
	return func(b *builder) {
		b.opts = opts
	}
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
		port: DefaultPort,
	}
	for _, opt := range opts {
		opt(b)
	}

	tcpServer := tcp.NewSubCommand("server", "s").
		Description("TCP/UDP Server Operations").
		Group("Test").
		Action(func(cmd *cmdr.Command, args []string) (err error) {
			return serverRun(cmd, args, b.opts...)
		})

	cmdr.NewBool().Titles("stop", "s", "shutdown").
		Description("stop/shutdown the running server").
		Group("Tool").
		AttachTo(tcpServer)

	tcpServer.NewFlagV(b.port, "port", "p").
		Description("The port to listen on").
		Group("TCP/UDP").
		Placeholder("PORT")

	cmdr.NewString().
		Titles("addr", "a", "adr", "address").
		Description("The address to listen to").
		Group("TCP/UDP").
		Placeholder("HOST-or-IP").
		AttachTo(tcpServer)

	cmdr.NewBool().Titles("0001.enable-tls", "tls").
		Description("enable TLS mode").
		Group("TLS").
		AttachTo(tcpServer)
	//tcpServer.NewFlagV(false, "enable-tls", "tls").
	//	Description("enable TLS mode").
	//	Group("TLS")

	cmdr.NewString("tcp").Titles("0007.network", "").
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
		Group("TLS").
		AttachTo(tcpServer)

	tcpServer.NewFlagV("root.pem", "100.cacert", "ca", "ca-cert").
		Description("CA cert path (.cer,.crt,.pem) if it's standalone").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.pem", "110.cert", "c").
		Description("server public-cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.key", "120.key", "k").
		Description("server private-key path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV(false, "190.client-auth").
		Description("enable client cert authentication").
		Group("TLS")
	tcpServer.NewFlagV(2, "200.tls-version").
		Description("tls-version: 0,1,2,3").
		Group("TLS")

	cmdr.NewString(DefaultPidPathTemplate).
		Titles("pid-path", "pp").
		Description("The pid filepath").
		Group("Tool").
		Placeholder("PATH").
		AttachTo(tcpServer)

}
