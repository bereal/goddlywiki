package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/markbates/pkger"
	"github.com/pkg/browser"
	"golang.org/x/net/webdav"
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

	src, err := pkger.Open("/data/empty.html")
	if err != nil {
		return
	}
	_, err = io.Copy(f, src)
	if err != nil {
		fmt.Printf("Create a new wiki file at %s", p)
	}
	return
}

type ServerConfig struct {
	HomeDir string
	Name    string
	File    string
	Port    int
}

func ConfigFromFile(f string, port int) ServerConfig {
	return ServerConfig{
		HomeDir: path.Dir(f),
		Name:    path.Base(f),
		File:    f,
		Port:    port,
	}
}

func ConfigFromName(name string, home string, port int) ServerConfig {
	if path.Ext(name) != ".html" {
		name = name + ".html"
	}

	return ServerConfig{
		HomeDir: home,
		Name:    name,
		File:    path.Join(home, name),
		Port:    port,
	}
}

func (s ServerConfig) URL() string {
	return fmt.Sprintf("http://localhost:%d/%s", s.Port, s.Name)
}

func (s ServerConfig) Init() error {
	if err := os.MkdirAll(s.HomeDir, 0775); err != nil {
		return err
	}

	if _, err := os.Stat(s.File); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err = CreateEmptyWiki(s.File)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s ServerConfig) Run() (func(), func()) {
	srv := &webdav.Handler{
		FileSystem: webdav.Dir(s.HomeDir),
		LockSystem: webdav.NewMemLS(),
	}

	addr := fmt.Sprintf(":%d", s.Port)
	hsrv := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := hsrv.ListenAndServe()
		if err != nil {
			die("Error starting the server: %s", err.Error())
		}
	}()

	return func() {
		fmt.Println("Shutting down...")
		hsrv.Shutdown(context.TODO())
	}, wg.Wait
}

var startCmd = flag.NewFlagSet("start", flag.ExitOnError)

var port = startCmd.Int("p", 8080, "port")
var home = startCmd.String("h", getBase(), "home directory")
var name = startCmd.String("n", "default", "wiki name")
var file = startCmd.String("f", "", "wiki file (overrides both -n and -h)")
var open = startCmd.Bool("o", false, "open in the browser")
var daemonize *bool

func init() {
	if runtime.GOOS != "windows" {
		daemonize = startCmd.Bool("d", false, "run as a daemon")
	} else {
		daemonize = new(bool)
	}
}

func start() {
	startCmd.Parse(os.Args[2:])
	var cfg ServerConfig
	if *file != "" {
		cfg = ConfigFromFile(*file, *port)
	} else {
		cfg = ConfigFromName(*name, *home, *port)
	}

	postStart := func() {
		if *open {
			<-time.After(time.Second)
			browser.OpenURL(cfg.URL())
		} else {
			fmt.Printf("The wiki is available at %s\n", cfg.URL())
		}
	}

	if err := cfg.Init(); err != nil {
		die("%s", err.Error())
	}

	if *daemonize {
		if runAsDaemon(cfg.Run) {
			postStart()
		}
	} else {
		cfg.Run()
		postStart()
		select {}
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
	default:
		die("Unexpected command: %s\n", os.Args[1])
	}
}
