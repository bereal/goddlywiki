package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
}

func ConfigFromFile(f string) ServerConfig {
	return ServerConfig{
		HomeDir: path.Dir(f),
		Name:    path.Base(f),
		File:    f,
	}
}

func ConfigFromName(name string, home string) ServerConfig {
	if path.Ext(name) != ".html" {
		name = name + ".html"
	}

	return ServerConfig{
		HomeDir: home,
		Name:    name,
		File:    path.Join(home, name),
	}
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

func (s ServerConfig) Run(port int) func() {
	srv := &webdav.Handler{
		FileSystem: webdav.Dir(s.HomeDir),
		LockSystem: webdav.NewMemLS(),
	}

	addr := fmt.Sprintf(":%d", port)
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
			log.Fatalf("Error starting the server: %s", err.Error())
		}
	}()

	return func() {
		fmt.Println("Shutting down...")
		hsrv.Shutdown(context.TODO())
		wg.Wait()
	}
}

func ensure(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func getHome() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}

var port = flag.Int("p", 8080, "port")
var home = flag.String("h", path.Join(getHome(), ".tiddly"), "home directory")
var name = flag.String("n", "default", "wiki name")
var file = flag.String("f", "", "wiki file (overrides both -n and -h)")
var open = flag.Bool("o", false, "open in the browser")

func main() {
	flag.Parse()
	var cfg ServerConfig
	if *file != "" {
		cfg = ConfigFromFile(*file)
	} else {
		cfg = ConfigFromName(*name, *home)
	}

	ensure(cfg.Init())
	url := fmt.Sprintf("http://localhost:%d/%s", *port, cfg.Name)
	fmt.Printf("Starting server, the wiki will be available at %s\n", url)
	cfg.Run(*port)

	if *open {
		<-time.After(time.Second)
		browser.OpenURL(url)
	}
	select {}
}
