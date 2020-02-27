package main

import (
	"context"
	"log"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

func main() {
	handler := NewHandler()

	var connOpt []jsonrpc2.ConnOpt

	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...,
	).DisconnectNotify()

	log.Println("golangci-lint-langserver: connections closed")
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
