package main

import (
	"os"
	"path"
	"sync"

	"github.com/sevlyar/go-daemon"
)

var daemonCtx *daemon.Context
var ctxOnce sync.Once

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
	go f()
	daemon.ServeSignals()
	return false
}
