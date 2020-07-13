package cert

import (
	"github.com/hedzr/cmdr"
	"log"
	"time"
)

func AttachToCmdr(optCmd cmdr.OptCmd) {
	certCmd := optCmd.NewSubCommand("cert").
		Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
		Group("Tool")

	certSubCommands(certCmd)
}

func certSubCommands(certOptCmd cmdr.OptCmd) {
	caCmd := certOptCmd.NewSubCommand("ca", "ca").
		Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
		Group("CA")

	caCreateCmd := caCmd.NewSubCommand("create", "c").
		Description("[NOT YET] create NEW CA certificates").
		Action(caCreate)
	log.Println(caCreateCmd)

	// certCmd := certOptCmd.NewSubCommand().
	// 	Titles("", "cert").
	// 	Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
	// 	Group("Tool")

	certCreateCmd := certOptCmd.NewSubCommand("create", "c").
		Description("create NEW certificates").
		Action(certCreate)
	log.Println(certCreateCmd)

	certCreateCmd.NewFlagV([]string{"localhost"}, "hostnames", "hns").
		Description("Comma-separated hostname list and/or IPs to generate a certificate for")
	certCreateCmd.NewFlagV("", "start-date", "sd", "from", "valid-from").
		Description("Creation date formatted as Jan 1 15:04:05 2011 (default now)")
	certCreateCmd.NewFlagV(365*10*24*time.Hour, "valid-for", "d", "duration").
		Description("Duration (10yr for debugging) that certificate is valid for")

	cmdr.NewString("./ci/certs").
		Titles("output-dir", "o", "output").
		Description("Creation date formatted as Jan 1 15:04:05 2011 (default now)").
		AttachTo(certCreateCmd)
}
