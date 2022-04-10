package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
)

func NewHandler(logger logger, noLinterName bool) jsonrpc2.Handler {
	handler := &langHandler{
		logger:       logger,
		request:      make(chan DocumentURI),
		noLinterName: noLinterName,
	}
	go handler.linter()

	return jsonrpc2.HandlerWithError(handler.handle)
}

type langHandler struct {
	logger       logger
	conn         *jsonrpc2.Conn
	request      chan DocumentURI
	command      []string
	noLinterName bool
	showAll      bool

	rootURI string
}

func (h *langHandler) lint(uri DocumentURI) (map[string][]Diagnostic, error) {
	diagnostics := make(map[string][]Diagnostic, 0)

	file := strings.TrimPrefix(string(uri), h.rootURI+string(os.PathSeparator))

	//nolint:gosec
	cmd := exec.Command(h.command[0], h.command[1:]...)
	cmd.Dir = uriToPath(h.rootURI)

	b, err := cmd.Output()
	if err == nil {
		return nil, nil
	}

	var result GolangCILintResult
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	h.logger.DebugJSON("golangci-lint-langserver: result:", result)

	for _, issue := range result.Issues {
		issue := issue

		if !h.showAll && file != issue.Pos.Filename {
			continue
		}

		uri := filepath.Join(h.rootURI, issue.Pos.Filename)

		//nolint:gomnd
		d := Diagnostic{
			Range: Range{
				Start: Position{Line: issue.Pos.Line - 1, Character: issue.Pos.Column - 1},
				End:   Position{Line: issue.Pos.Line - 1, Character: issue.Pos.Column - 1},
			},
			Severity: DSWarning,
			Source:   &issue.FromLinter,
			Message:  h.diagnosticMessage(&issue),
		}

		if _, ok := diagnostics[uri]; !ok {
			diagnostics[uri] = []Diagnostic{d}
		} else {
			diagnostics[uri] = append(diagnostics[uri], d)
		}
	}

	return diagnostics, nil
}

func (h *langHandler) diagnosticMessage(issue *Issue) string {
	if h.noLinterName {
		return issue.Text
	}

	return fmt.Sprintf("%s: %s", issue.FromLinter, issue.Text)
}

func (h *langHandler) linter() {
	for {
		uri, ok := <-h.request
		if !ok {
			break
		}

		diagnosticsMap, err := h.lint(uri)
		if err != nil {
			h.logger.Printf("%s", err)

			continue
		}

		for uri, diagnostics := range diagnosticsMap {
			if err := h.conn.Notify(
				context.Background(),
				"textDocument/publishDiagnostics",
				&PublishDiagnosticsParams{
					URI:         DocumentURI(uri),
					Diagnostics: diagnostics,
				},
			); err != nil {
				h.logger.Printf("%s", err)
			}
		}
	}
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
	case "textDocument/didOpen":
		return h.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return h.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		return h.handleTextDocumentDidSave(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *langHandler) handleInitialize(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.rootURI = params.RootURI
	h.conn = conn
	h.showAll = params.InitializationOptions.ShowAll
	h.command = params.InitializationOptions.Command
	// default command if none is specified
	if len(h.command) == 0 {
		h.command = []string{"golangci-lint", "run", "--out-format", "json"}
	}

	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TextDocumentSyncOptions{
				Change:    TDSKNone,
				OpenClose: true,
				Save:      true,
			},
		},
	}, nil
}

func (h *langHandler) handleShutdown(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	close(h.request)

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidOpen(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidClose(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidChange(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidSave(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}
