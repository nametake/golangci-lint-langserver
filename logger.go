package main

import (
	"encoding/json"
	"log"
	"os"
)

var _ logger = (*stdLogger)(nil)

type logger interface {
	Printf(format string, args ...interface{})
	DebugJSON(label string, arg interface{})
}

type stdLogger struct {
	debug  bool
	stderr *log.Logger
}

func newStdLogger(debug bool) *stdLogger {
	return &stdLogger{
		debug:  debug,
		stderr: log.New(os.Stderr, "", 0),
	}
}

func (l *stdLogger) Printf(format string, args ...interface{}) {
	l.stderr.Printf(format, args...)
}

func (l *stdLogger) DebugJSON(label string, arg interface{}) {
	if !l.debug {
		return
	}

	b, err := json.Marshal(arg)
	if err != nil {
		l.stderr.Println(err)
	}

	l.stderr.Println(label, string(b))
}
