module github.com/hedzr/go-socketlib/_examples

go 1.21

//replace github.com/hedzr/is => ../../../cmdr.v2/libs.is

//replace github.com/hedzr/env => ../../../cmdr.v2/libs.env

//replace github.com/hedzr/logg => ../../../cmdr.v2/libs.logg

replace github.com/hedzr/go-socketlib => ../

require (
	github.com/hedzr/go-socketlib v1.0.1
	github.com/hedzr/is v0.5.17
	github.com/miekg/dns v1.1.58
)

require (
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
)
