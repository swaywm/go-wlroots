package main

import (
	"flag"
	"os"
	"os/exec"
)

var (
	command = flag.String("s", "", "startup command")
)

func main() {
	flag.Parse()

	// start the server
	server, err := NewServer()
	if err != nil {
		panic(err)
	}
	if err = server.Start(); err != nil {
		panic(err)
	}

	// run the startup command if given
	if *command != "" {
		cmd := exec.Command("/bin/sh", "-c", *command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Start(); err != nil {
			panic(err)
		}
	}

	// start the wayland event loop
	if err = server.Run(); err != nil {
		panic(err)
	}
}
