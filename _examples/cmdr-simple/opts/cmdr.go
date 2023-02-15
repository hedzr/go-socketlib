package opts

import (
	"github.com/hedzr/cmdr"

	"github.com/hedzr/go-socketlib/tcp/cert"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"
)

func AttachToCmdr(root cmdr.OptCmd) {
	socketLibCmd(root)
}

func socketLibCmd(root cmdr.OptCmd) {

	// TCP/UDP

	tcpCmd := cmdr.NewSubCmd().Titles("tcp", "tcp", "socketlib").
		Description("go-socketlib TCP operations...", "").
		Group("Socket").
		AttachTo(root)

	server.AttachToCmdrCommand(tcpCmd, server.WithCmdrPort(1983))
	client.AttachToCmdrCommand(tcpCmd, client.WithCmdrPort(1983), client.WithCmdrInteractiveCommand(true))

	udpCmd := cmdr.NewSubCmd().Titles("udp", "udp").
		Description("go-socketlib UDP operations...", "").
		Group("Socket").
		AttachTo(root)

	server.AttachToCmdrCommand(udpCmd, server.WithCmdrUDPMode(true), server.WithCmdrPort(1984))
	client.AttachToCmdrCommand(udpCmd, client.WithCmdrUDPMode(true), client.WithCmdrPort(1984))

	// Cert

	cert.AttachToCmdrCommand(root)
}
