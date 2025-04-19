package main

import (
	"context"
	"flag"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

var defaultSeverity = "Warn"

func main() {
	debug := flag.Bool("debug", false, "output debug log")
	noLinterName := flag.Bool("nolintername", false, "don't show a linter name in message")
	flag.StringVar(&defaultSeverity, "severity", defaultSeverity, "Default severity to use. Choices are: Err(or), Warn(ing), Info(rmation) or Hint")

	flag.Parse()

	logger := newStdLogger(*debug)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	handler := NewHandler(&HandlerConfig{
		Logger:       logger,
		NoLinterName: *noLinterName,
		Close:        cancel,
	})

	var connOpt []jsonrpc2.ConnOpt
	if *debug {
		connOpt = append(connOpt, jsonrpc2.LogMessages(logger))
	}

	conn := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...,
	)

	defer conn.Close()

	logger.Printf("golangci-lint-langserver: connection opened")
	defer logger.Printf("golangci-lint-langserver: connection closed")

	select {
	case <-ctx.Done():
	case <-conn.DisconnectNotify():
	}
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}

	return os.Stdout.Close()
}
