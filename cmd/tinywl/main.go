package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/swaywm/go-wlroots/wlroots"
)

var (
	command      = flag.String("s", "", "startup command")
	programLevel = new(slog.LevelVar) // Info by default

)

func fatal(msg string, err error) {
	fmt.Printf("error %s: %s\n", msg, err)
	os.Exit(1)
}

func init() {
	// lock the main goroutine onto the current OS thread
	// we need to do this because EGL uses thread local storage
	runtime.LockOSThread()
}

func main() {
	// set global logger with custom options
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})))

	flag.Parse()

	// set up logging
	// programLevel.Set(slog.LevelDebug)
	wlroots.OnLog(wlroots.LogImportanceDebug, nil)

	// start the server
	server, err := NewServer()
	if err != nil {
		fatal("initializing server", err)
	}
	if err = server.Start(); err != nil {
		fatal("starting server", err)
	}

	// run the startup command if given
	if *command != "" {
		cmd := exec.Command("/bin/sh", "-c", *command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Start(); err != nil {
			fatal("running startup command", err)
		}
	}

	// start the wayland event loop
	if err = server.Run(); err != nil {
		fatal("running server", err)
	}
}
