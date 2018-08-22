package main

import (
	"bytes"
	logger "log"
	"os"

	"github.com/alexbakker/go-wlroots/wlroots"
	isatty "github.com/mattn/go-isatty"
)

var (
	log          = logger.New(os.Stderr, "", logger.Ldate|logger.Lmicroseconds)
	logVerbosity = wlroots.LogImportanceDebug

	logColors = map[wlroots.LogImportance]string{
		wlroots.LogImportanceSilent: "",
		wlroots.LogImportanceError:  "\x1B[1;31m",
		wlroots.LogImportanceInfo:   "\x1B[1;34m",
		wlroots.LogImportanceDebug:  "\x1B[1;30m",
	}
)

func handleLog(importance wlroots.LogImportance, msg string) {
	isTerm := isatty.IsTerminal(os.Stderr.Fd())
	if importance > logVerbosity {
		return
	}

	var buf bytes.Buffer
	if isTerm {
		buf.WriteString(logColors[importance])
	}
	buf.WriteString(msg)
	if isTerm {
		buf.WriteString("\x1B[0m")
	}

	log.Println(buf.String())
}

func init() {
	wlroots.OnLog(logVerbosity, handleLog)
}
