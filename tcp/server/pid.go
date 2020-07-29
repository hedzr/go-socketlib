package server

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex/sig"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log/exec"
	"gopkg.in/hedzr/errors.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

func makePidFS(prefixInCommandLine string) *pidFileStruct {
	return newPidFile(cmdr.GetStringRP(prefixInCommandLine, "pid-path", base.DefaultPidPathTemplate))
}

func makePidFSFromDir(dir string) *pidFileStruct {
	pp := path.Join(dir, "$APPNAME.pid")
	return newPidFile(pp)
}

func findAndShutdownTheRunningInstance(pfs *pidFileStruct) (err error) {
	var present bool
	var process *os.Process
	if present, process, err = findDaemonProcess(pfs); err == nil && present {
		err = sig.SendQUIT(process)
	}
	return
}

type pidFileStruct struct {
	Path string
}

// var pidfile = &pidFileStruct{}

func newPidFile(filepath string) *pidFileStruct {
	return &pidFileStruct{
		Path: os.ExpandEnv(filepath),
	}
}

func (pf *pidFileStruct) String() string {
	return pf.Path
}

func (pf *pidFileStruct) Create() (err error) {
	if err = exec.EnsureDirEnh(path.Dir(pf.Path)); err != nil {
		return
	}

	var f *os.File
	f, err = os.OpenFile(pf.Path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0770)
	if err != nil {
		err = errors.New("Failed to create pid file %q", pf.Path).Attach(err)
	}

	defer func() { err = f.Close() }()

	_, err = f.WriteString(fmt.Sprintf("%v", os.Getpid()))
	return
}

func (pf *pidFileStruct) Destroy() {
	// if cmdr.GetBoolR("server.start.in-daemon") {
	//	//
	// }
	if cmdr.FileExists(pf.Path) {
		err := os.RemoveAll(pf.Path)
		if err != nil {
			panic(errors.New("Failed to destroy pid file %q", pf.Path).Attach(err))
		}
	}
}

func pidExistsDeep(pid int) (bool, error) {
	// pid, err := strconv.ParseInt(p, 0, 64)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	process, err := os.FindProcess(int(pid))
	if err != nil {
		// fmt.Printf("Failed to find process: %s\n", err)
		return false, nil
	}

	err = sig.SendNilSig(process)
	log.Printf("process.Signal on pid %d returned: %v\n", pid, err)
	return err == nil, err
}

// isPidFileExists checks if the pid file exists or not
func isPidFileExists(pfs *pidFileStruct) bool {
	// check if daemon already running.
	if _, err := os.Stat(pfs.Path); err == nil {
		return true

	}
	return false
}

// findDaemonProcess locates the daemon process if running
func findDaemonProcess(pfs *pidFileStruct) (present bool, process *os.Process, err error) {
	if isPidFileExists(pfs) {
		s, _ := ioutil.ReadFile(pfs.Path)
		var pid int64
		pid, err = strconv.ParseInt(string(s), 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("cat %v ... pid = %v", pfs, pid)

		process, err = os.FindProcess(int(pid))
		if err == nil {
			present = true
		}
	} else {
		err = errors.New("cat %v ... app stopped", pfs.Path)
	}
	return
}

const nullDev = "/dev/null"
