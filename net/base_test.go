package net

import (
	"os"
	"testing"
)

func TestBaseS_error(t *testing.T) {
	var b = newBaseS()
	b.Error("error msg")
	t.Log("")
}

type redisHub struct{}

func (s *redisHub) Close() {
	// close the connections to redis servers
	println("redis connections closed")
}

func TestBaseS_closers(t *testing.T) {
	base := newBaseS()

	defer base.Close()

	base.addClosable(&redisHub{})

	base.addCloseFunc(func() {
		// do some shutdown operations here
		println("close functor")
	})

	base.addCloseFunc(func() {
		// do some shutdown operations here
		println("close single functor")
	})

	tmpFile, err := os.CreateTemp(os.TempDir(), "1*.log")
	t.Logf("tmpfile: %v | err: %v", tmpFile.Name(), err)
	base.addCloser(tmpFile)

	for _, ii := range base.closers {
		println(ii)
	}

	// These following calls are both unused since we
	// have had a defer basics.Close().
	// But they are harmless here.
}
