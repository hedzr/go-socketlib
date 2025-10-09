module github.com/hedzr/go-socketlib/_examples

go 1.24.0

toolchain go1.24.5

//replace github.com/hedzr/is => ../../../cmdr.v2/libs.is

//replace github.com/hedzr/env => ../../../cmdr.v2/libs.env

//replace github.com/hedzr/logg => ../../../cmdr.v2/libs.logg

replace github.com/hedzr/go-socketlib => ../

require (
	github.com/hedzr/go-socketlib v1.1.7
	github.com/hedzr/is v0.8.60
	github.com/miekg/dns v1.1.68
)

require (
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
)
