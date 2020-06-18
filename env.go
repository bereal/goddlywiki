package main

import (
	"fmt"
	"os"
	"path"
	"sync"
)

var basedir string
var basedirOnce sync.Once

func getBase() string {
	basedirOnce.Do(func() {
		envHome := os.Getenv("GODDLY_HOME")
		if envHome != "" {
			basedir = envHome
		} else {
			userHome, _ := os.UserHomeDir()
			basedir = path.Join(userHome, ".tiddly")
		}
	})

	return basedir
}

func die(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(1)
}
