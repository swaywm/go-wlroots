package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/alexbakker/go-wlroots/wlroots"
)

var (
	command = flag.String("s", "", "startup command")
)

func main() {
	flag.Parse()

	// set up logging
	wlroots.OnLog(wlroots.LogImportanceDebug, nil)

	// start the server
	server, err := NewServer()
	if err != nil {
		fmt.Printf("error creating server: %s\n", err)
	}
	if err = server.Start(); err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}

	// run the startup command if given
	if *command != "" {
		cmd := exec.Command("/bin/sh", "-c", *command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Start(); err != nil {
			fmt.Printf("error running startup command: %s\n", err)
		}
	}

	// start the wayland event loop
	if err = server.Run(); err != nil {
		fmt.Printf("error running server: %s\n", err)
	}
}
