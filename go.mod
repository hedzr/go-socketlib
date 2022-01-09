module github.com/hedzr/go-socketlib

go 1.13

//replace github.com/hedzr/cmdr => ../../go-cmdr/50.cmdr

// replace github.com/hedzr/log => ../../go-cmdr/10.log

// replace github.com/hedzr/logex => ../../go-cmdr/13.logex

// replace github.com/hedzr/go-ringbuf => ../go-ringbuf

// replace github.com/hedzr/rules => ../rules

// replace github.com/hedzr/pools => ../pools

// replace github.com/hedzr/errors => ../errors

require (
	github.com/hedzr/cmdr v1.9.7
	github.com/hedzr/cmdr-addons v1.9.7
	github.com/hedzr/go-ringbuf v0.8.9
	github.com/hedzr/log v1.3.23
	github.com/hedzr/logex v1.3.23
	gopkg.in/hedzr/errors.v2 v2.1.5
)
