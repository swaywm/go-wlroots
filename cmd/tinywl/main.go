package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// start the server
	server, err := NewServer()
	if err != nil {
		panic(err)
	}
	if err = server.Start(); err != nil {
		panic(err)
	}

	// run the requested command
	cmd := exec.Command("thunar")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		panic(err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("exec error: %s\n", err)
		}
	}()

	// start the wayland event loop
	if err = server.Run(); err != nil {
		panic(err)
	}
}
