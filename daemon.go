package main

import (
	"os"
	"path"
	"sync"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

var daemonCtx *daemon.Context
var ctxOnce sync.Once

var stopService func()
var waitService func()

func getDaemonContext() *daemon.Context {
	ctxOnce.Do(func() {
		base := getBase()
		err := os.MkdirAll(base, 0755)
		if err != nil {
			die("%s\n", err.Error())
		}

		daemonCtx = &daemon.Context{
			PidFileName: path.Join(base, "goddly.pid"),
			PidFilePerm: 0644,
			LogFileName: path.Join(base, "goddly.log"),
			LogFilePerm: 0640,
			Umask:       027,
		}
	})
	return daemonCtx
}

type launchFunc func() (func(), func())

func runAsDaemon(f launchFunc) bool {
	ctx := getDaemonContext()
	d, err := ctx.Reborn()
	if err != nil {
		die("%s\n", err.Error())
	}

	if d != nil {
		return true
	}

	defer ctx.Release()
	stopService, waitService = f()
	daemon.SetSigHandler(func(sig os.Signal) error {
		stopService()
		if sig == syscall.SIGQUIT {
			waitService()
		}
		return daemon.ErrStop
	}, syscall.SIGQUIT, syscall.SIGTERM)
	daemon.ServeSignals()
	return false
}

func stopDaemon() error {
	ctx := getDaemonContext()
	pid, err := daemon.ReadPidFile(ctx.PidFileName)
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return proc.Signal(syscall.SIGQUIT)
}
