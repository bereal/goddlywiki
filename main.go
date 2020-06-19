package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/markbates/pkger"
	"github.com/pkg/browser"
)

func CreateEmptyWiki(p string) (err error) {
	f, err := os.Create(p)
	if err != nil {
		return
	}

	defer func() {
		err1 := f.Close()
		if err == nil {
			err = err1
		}
	}()

	gz, err := pkger.Open("/data/empty.gz")
	if err != nil {
		return
	}

	src, err := gzip.NewReader(gz)
	if err != nil {
		return
	}

	_, err = io.Copy(f, src)
	if err == nil {
		fmt.Printf("Created a new wiki file at %s\n", p)
	}
	return
}

var startCmd = flag.NewFlagSet("start", flag.ExitOnError)

var startFlags struct {
	port      int
	home      string
	name      string
	file      string
	open      bool
	daemonize bool
}

var createCmd = flag.NewFlagSet("create", flag.ExitOnError)

var createFlags struct {
	home string
	name string
	file string
}

func init() {
	startCmd.IntVar(&startFlags.port, "p", 8080, "port")
	startCmd.StringVar(&startFlags.home, "h", getBase(), "home directory")
	startCmd.StringVar(&startFlags.name, "n", "default", "wiki name")
	startCmd.StringVar(&startFlags.file, "f", "", "wiki file (overrides both -n and -h)")
	startCmd.BoolVar(&startFlags.open, "o", false, "open in the browser")

	if runtime.GOOS != "windows" {
		startCmd.BoolVar(&startFlags.daemonize, "d", false, "run as a daemon")
	}

	createCmd.StringVar(&createFlags.home, "h", getBase(), "home directory")
	createCmd.StringVar(&createFlags.name, "n", "default", "wiki name")
	createCmd.StringVar(&createFlags.file, "f", "", "wiki file (overrides both -n and -h)")
}

func start() {
	startCmd.Parse(os.Args[2:])

	var cfg ServerConfig
	if startFlags.file != "" {
		cfg = ConfigFromFile(startFlags.file, startFlags.port)
	} else {
		cfg = ConfigFromName(startFlags.name, startFlags.home, startFlags.port)
	}

	postStart := func() {
		if startFlags.open {
			<-time.After(time.Second)
			browser.OpenURL(cfg.URL())
		} else {
			fmt.Printf("The wiki is available at %s\n", cfg.URL())
		}
	}

	if err := cfg.Init(); err != nil {
		die("%s", err.Error())
	}

	if startFlags.daemonize {
		if runAsDaemon(cfg.Run) {
			postStart()
		}
	} else {
		cfg.Run()
		postStart()
		select {}
	}
}

func create() {
	createCmd.Parse(os.Args[2:])

	var cfg ServerConfig
	if createFlags.file != "" {
		cfg = ConfigFromFile(createFlags.file, 0)
	} else {
		cfg = ConfigFromName(createFlags.name, createFlags.home, 0)
	}

	if err := CreateEmptyWiki(cfg.File); err != nil {
		die("Error creating wiki file: %s\n", err.Error())
	}
}

func main() {
	if len(os.Args) < 2 {
		die("Expected a command\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		start()
	case "stop":
		if err := stopDaemon(); err != nil {
			die("Error stopping: %s\n", err.Error())
		}
	case "create":
		create()
	default:
		die("Unexpected command: %s\n", os.Args[1])
	}
}
