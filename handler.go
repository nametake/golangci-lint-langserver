package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
)

func NewHandler() jsonrpc2.Handler {
	handler := &langHandler{}
	return jsonrpc2.HandlerWithError(handler.handle)
}

type langHandler struct {
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
