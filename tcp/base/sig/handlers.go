// Copyright © 2020 Hedzr Yeh.

package sig

import (
	"log"
	"net"
	"os"
	"os/exec"

	"gopkg.in/hedzr/errors.v3"
)

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	// if daemonImpl != nil {
	// 	daemonImpl.OnReload()
	// }
	return nil
}

func hotReloadHandler() func(sig os.Signal) error {
	return func(sig os.Signal) error {
		log.Println("hot-reloaded")

		if onGetListener != nil {
			listener := onGetListener()

			tl, ok := listener.(*net.TCPListener)
			if !ok {
				return errors.New("listener is not tcp listener")
			}

			f, err := tl.File()
			if err != nil {
				return err
			}

			log.Printf("f: %v", f.Name())
			log.Printf("bin: %v", os.Args[0])
			args := []string{"server", "start", "--in-hot-reload"}
			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout         //
			cmd.Stderr = os.Stderr         //
			cmd.ExtraFiles = []*os.File{f} //
			if err = cmd.Start(); err != nil {
				return err
			}
		}

		// if ctx.onHotReloading != nil {
		// 	return ctx.onHotReloading(ctx)
		// }
		return nil
	}
}
