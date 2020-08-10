package message

import (
	"github.com/hedzr/log"
	"sync"
)

var (
	once   sync.Once
	logger log.Logger
)

func init() {
	once.Do(func() {
		// instance = make(Config)
		logger = log.NewStdLogger()
	})
}

func SetLogger(l log.Logger) {
	logger = l
}

func Logger() log.Logger {
	return logger
}
