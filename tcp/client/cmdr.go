/*
 * Copyright © 2020 Hedzr Yeh.
 */

package client

import (
	"github.com/hedzr/cmdr"
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

func WithCmdrInteractiveCommand(enabled bool) CmdrOpt {
	return func(b *builder) {
		b.interactiveCommand = enabled
	}
}

func WithCmdrCommandAction(action cmdr.Handler) CmdrOpt {
	return func(b *builder) {
		b.action = action
	}
}

type builder struct {
	port               int
	interactiveCommand bool
	action             cmdr.Handler
	udpMode            bool
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
	}
	for _, opt := range opts {
		opt(b)
	}

	if b.interactiveCommand {
		tc2 := tcp.NewSubCommand("interactive-client", "ic").
			Description("TCP interactive client operations").
			Group("Test").
			Action(interactiveRunAsCliTool)
		b.attachTcpClientFlags(tc2)
	}

	theClient := tcp.NewSubCommand("client", "c").
		Description("TCP client operations").
		Group("Test").
		Action(b.action)

	b.attachTcpClientFlags(theClient)

	if !b.udpMode {
		b.attachTcpTLSClientFlags(theClient)
	}
}

func (b *builder) attachTcpClientFlags(theClient cmdr.OptCmd) {

	network := "tcp"
	if b.udpMode {
		// b.opts = append(b.opts, WithServerUDPMode(true))
		network = "udp"
	}

	theClient.NewFlagV(b.port, "port", "p").
		Description("The port to connect to").
		Group("Test").
		Placeholder("PORT")

	theClient.NewFlagV("127.0.0.1", "host", "h", "address", "addr").
		Description("The hostname or IP to connect to").
		Group("Test").
		Placeholder("HOST-or-IP")
	// don't use localhost, it may cause 'lookup localhost: no such host' error in debug mode.

	theClient.NewFlagV(100, "times", "t").
		Description("repeat sending times").
		Group("Test").
		Placeholder("n")

	theClient.NewFlagV(3, "parallel", "r").
		Description("how many clients parallel").
		Group("Test").
		Placeholder("n")

	theClient.NewFlagV(time.Duration(0), "sleep").
		Description("sleep time between each sending").
		Group("Test")

	theClient.NewFlagV(false, "interactive", "i").
		Description("run client in interactive mode").
		Group("Test")

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
		Group("TLS").
		AttachTo(theClient)

}

func (b *builder) attachTcpTLSClientFlags(theClient cmdr.OptCmd) {
	theClient.NewFlagV(false, "enable-tls", "tls").
		Description("enable TLS mode").
		Group("TLS")

	theClient.NewFlagV("root.pem", "cacert", "ca").
		Description("CA cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	theClient.NewFlagV("cert.pem", "server-cert", "sc").
		Description("server public-cert path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	theClient.NewFlagV("client.pem", "cert").
		Description("[client-auth] client public-cert path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	theClient.NewFlagV("client.key", "key").
		Description("[client-auth] client private-key path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	theClient.NewFlagV(false, "client-auth").
		Description("[client-auth] enable client cert authentication").
		Group("TLS")
	theClient.NewFlagV(false, "insecure", "k").
		Description("[client-auth] ignore server cert validation (for self-signed server)").
		Group("TLS")
	theClient.NewFlagV(2, "tls-version").
		Description("tls-version: 0,1,2,3").
		Group("TLS")

}
