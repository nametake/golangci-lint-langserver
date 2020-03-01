package main

import (
	"context"
	"encoding/json"
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
	logger  logger
	rootURI string
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	h.logger.DebugJSON("golangci-lint-langserver: request:", req)

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

func (h *langHandler) handleInitialize(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	params := &InitializeParams{}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.rootURI = params.RootURI

	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TDSKFull,
		},
	}, nil
}

func (h *langHandler) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}
