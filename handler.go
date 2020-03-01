package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
)

func NewHandler(logger logger) jsonrpc2.Handler {
	handler := &langHandler{
		logger: logger,
	}
	return jsonrpc2.HandlerWithError(handler.handle)
}

type langHandler struct {
	logger logger
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	h.logger.DebugJSON("golangci-lint-langserver: handler:", req)
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "initialized":
		return
	case "shutdown":
		return h.handleShutdown(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *langHandler) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return InitializeResult{
		Capabilities: ServerCapabilities{},
	}, nil
}

func (h *langHandler) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}
