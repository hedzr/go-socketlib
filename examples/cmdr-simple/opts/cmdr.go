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

	tcpCmd := root.NewSubCommand("tcp", "tcp", "socketlib").
		Description("go-socketlib TCO operations...", "").
		Group("Socket")

	server.AttachToCmdr(tcpCmd, server.WithCmdrPort(1983))
	client.AttachToCmdr(tcpCmd, client.WithCmdrPort(1983), client.WithCmdrInteractiveCommand(true))

	udpCmd := root.NewSubCommand("udp", "udp").
		Description("go-socketlib UDP operations...", "").
		Group("Socket")

	server.AttachToCmdr(udpCmd, server.WithCmdrUDPMode(true), server.WithCmdrPort(1984))
	client.AttachToCmdr(udpCmd, client.WithCmdrUDPMode(true), client.WithCmdrPort(1984))

	// Cert

	cert.AttachToCmdr(root)

}
