module github.com/hedzr/go-socketlib

go 1.13

replace github.com/hedzr/cmdr => ../cmdr

// replace github.com/hedzr/log => ../log

// replace github.com/hedzr/logex => ../logex

// replace github.com/hedzr/go-ringbuf => ../go-ringbuf

// replace github.com/hedzr/rules => ../rules

// replace github.com/hedzr/pools => ../pools

// replace github.com/hedzr/errors => ../errors

require (
	github.com/hedzr/cmdr v1.7.9
	github.com/hedzr/cmdr-addons v1.7.9
	github.com/hedzr/go-ringbuf v0.8.7
	github.com/hedzr/log v0.1.16
	github.com/hedzr/logex v1.2.8
	gopkg.in/hedzr/errors.v2 v2.0.12
)
