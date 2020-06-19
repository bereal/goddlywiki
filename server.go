package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"

	"golang.org/x/net/webdav"
)

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
	dav := &webdav.Handler{
		FileSystem: webdav.Dir(s.HomeDir),
		LockSystem: webdav.NewMemLS(),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			dav.ServeHTTP(w, r)
		} else {
			w.Header().Add("Location", "/"+s.Name)
			w.WriteHeader(http.StatusFound)
		}
	})

	addr := fmt.Sprintf(":%d", s.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := srv.ListenAndServe()
		if err != nil {
			die("Error starting the server: %s", err.Error())
		}
	}()

	return func() {
		fmt.Println("Shutting down...")
		srv.Shutdown(context.TODO())
	}, wg.Wait
}
