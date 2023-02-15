package base

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/hedzr/log/dir"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/go-socketlib/tcp/base/sig"
)

func makePidFS(prefixInCommandLine, prefixInConfigFile, defaultDir string) *pidFileStruct {
	var dPath string
	if defaultDir == "" {
		dPath = DefaultPidPathTemplate
	} else {
		dPath = path.Join(defaultDir, "$APPNAME.pid")
	}

	var str string
	str = cmdr.GetStringRP(prefixInCommandLine, "pid-path", "")
	if str == "" {
		str = cmdr.GetStringRP(prefixInCommandLine, "pid-path", "")
	}
	if str == "" {
		str = os.ExpandEnv(dPath)
	}
	return newPidFile(str)
}

func makePidFSFromDir(dir string) *pidFileStruct {
	pp := path.Join(dir, "$APPNAME.pid")
	return newPidFile(pp)
}

type PidFile interface {
	Path() string
	Create(baseCtx context.Context) (err error)
	Destroy()
	IsExists() bool
}

type pidFileStruct struct {
	path string
}

// var pidfile = &pidFileStruct{}

func newPidFile(filepath string) *pidFileStruct {
	return &pidFileStruct{
		path: os.ExpandEnv(filepath),
	}
}

func (pf *pidFileStruct) String() string {
	return pf.path
}

func (pf *pidFileStruct) Path() string {
	return pf.path
}

func (pf *pidFileStruct) Create(ctx context.Context) (err error) {
	d := path.Dir(pf.path)
	if err = dir.EnsureDir(d); err != nil {
		fmt.Printf(`

You're been prompt with a "sudo" requesting because this folder was been creating but need more privileges:

- %v

We must have created a PID file in it.

`, d)
		if err = dir.EnsureDirEnh(d); err != nil {
			return
		}
	}

	var f *os.File
	f, err = os.OpenFile(pf.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0770)
	if err != nil {
		err = errors.New("Failed to create pid file %q", pf.path).WithErrors(err)
	}

	defer func() { err = f.Close() }()

	_, err = f.WriteString(fmt.Sprintf("%v", os.Getpid()))
	return
}

func (pf *pidFileStruct) Destroy() {
	// if cmdr.GetBoolR("server.start.in-daemon") {
	//	//
	// }
	if dir.FileExists(pf.path) {
		err := os.RemoveAll(pf.path)
		if err != nil {
			panic(errors.New("Failed to destroy pid file %q", pf.Path).WithErrors(err))
		}
		log.Infof("%q destroyed", pf.path)
	}
}

func (pf *pidFileStruct) IsExists() bool {
	// check if daemon already running.
	if _, err := os.Stat(pf.path); err == nil {
		return true
	}
	return false
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

// IsPidFileExists checks if the pid file exists or not
func IsPidFileExists(pfs PidFile) bool {
	return pfs.IsExists()
}

// FindDaemonProcess locates the daemon process if running
func FindDaemonProcess(pfs PidFile) (present bool, process *os.Process, err error) {
	if IsPidFileExists(pfs) {
		s, _ := ioutil.ReadFile(pfs.Path())
		var pid int64
		pid, err = strconv.ParseInt(string(s), 0, 64)
		if err != nil {
			log.Fatalf("%v", err)
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

// FindAndShutdownTheRunningInstance locates the daemon process if running
func FindAndShutdownTheRunningInstance(pfs PidFile) (err error) {
	var present bool
	var process *os.Process
	if present, process, err = FindDaemonProcess(pfs); err == nil && present {
		err = sig.SendQUIT(process)
	}
	return
}

const nullDev = "/dev/null"
