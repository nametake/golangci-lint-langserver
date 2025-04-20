package main

import (
	"encoding/json"
	"log"
)

var _ logger = (*stdLogger)(nil)

type logger interface {
	Printf(format string, args ...any)
	DebugJSON(label string, arg any)
}

type stdLogger struct {
	debug  bool
	stderr *log.Logger
}

func newStdLogger(debug bool) *stdLogger {
	return &stdLogger{
		debug:  debug,
		stderr: log.New(log.Writer(), "", log.LstdFlags),
	}
}

func (l *stdLogger) Printf(format string, args ...any) {
	l.stderr.Printf(format, args...)
}

func (l *stdLogger) DebugJSON(label string, arg any) {
	if !l.debug {
		return
	}

	b, err := json.Marshal(arg)
	if err != nil {
		l.stderr.Println(err)
		return
	}

	l.stderr.Println(label, string(b))
}
