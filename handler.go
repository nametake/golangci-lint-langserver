package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/jsonrpc2"
)

func NewHandler(logger logger, noLinterName bool, fromConfigDir bool) jsonrpc2.Handler {
	handler := &langHandler{
		logger:        logger,
		request:       make(chan DocumentURI),
		noLinterName:  noLinterName,
		fromConfigDir: fromConfigDir,
	}
	go handler.linter()

	return jsonrpc2.HandlerWithError(handler.handle)
}

type langHandler struct {
	logger        logger
	conn          *jsonrpc2.Conn
	request       chan DocumentURI
	command       []string
	noLinterName  bool
	fromConfigDir bool

	rootURI string
}

func (h *langHandler) findConfigDirectory(file string) (string, error) {
	dir, _ := filepath.Split(file)

	// Maybe this should be configurable too?
	findConfigCmd := exec.Command(h.command[0], "config", "path")
	findConfigCmd.Dir = dir

	configFile, err := findConfigCmd.Output()
	if err != nil {
		return "", err
	}

	configDirRel, _ := filepath.Split(string(configFile))

	// configFileBytes is relative to dir, we need to make it absolute
	//
	// This assumes that `dir` is absolute
	return filepath.Join(dir, configDirRel), nil
}

func (h *langHandler) runLinter(dir string) (GolangCILintResult, error) {
	//nolint:gosec
	cmd := exec.Command(h.command[0], h.command[1:]...)

	// linter might be ran either from the directory of the file or from the directory of the config file
	var checkDir string
	if h.fromConfigDir {
		// Relative paths in .golangci-lint.yml work, but
		// - we need to find the directory with config
		// - we have to adjust the paths in the result

		configDir, err := h.findConfigDirectory(dir)
		if err != nil {
			return GolangCILintResult{}, err
		}

		h.logger.Printf("Found golangci-lint config file in directory %s", configDir)

		// Find the original directory, relative to the config dir
		checkDir, err = filepath.Rel(configDir, dir)
		if err != nil {
			return GolangCILintResult{}, err
		}

		// This runs the linter on the subtree. Non-config-dir option does not
		// pass any packages to the command line, which is equivalent.
		cmd.Args = append(cmd.Args, checkDir+"/...")
		cmd.Dir = configDir
	} else {
		// Relative paths in golangci-lint.yml don't work, but the paths in report are correct,
		// and no additional work is needed
		cmd.Dir = dir
	}

	b, err := cmd.Output()
	if err == nil { // This expects that the golangci-lint exits with non-zero code on errors
		return GolangCILintResult{}, nil
	}

	var result GolangCILintResult
	if err := json.Unmarshal(b, &result); err != nil {
		return GolangCILintResult{}, err
	}

	// We need to adjust the paths in the result (see above)
	if h.fromConfigDir {
		var issues []Issue
		for _, issue := range result.Issues {
			// Strip checkDir from the path
			issue.Pos.Filename, err = filepath.Rel(checkDir, issue.Pos.Filename)
			if err != nil {
				return GolangCILintResult{}, err
			}
			issues = append(issues, issue)
		}
		result.Issues = issues
	}

	return result, nil
}

func (h *langHandler) lint(uri DocumentURI) ([]Diagnostic, error) {
	dir, file := filepath.Split(uriToPath(string(uri)))

	diagnostics := make([]Diagnostic, 0)

	result, err := h.runLinter(dir)
	if err != nil {
		return diagnostics, err
	}

	h.logger.DebugJSON("golangci-lint-langserver: result:", result)

	for _, issue := range result.Issues {
		issue := issue

		if file != issue.Pos.Filename {
			continue
		}

		d := Diagnostic{
			Range: Range{
				Start: Position{Line: issue.Pos.Line - 1, Character: issue.Pos.Column - 1},
				End:   Position{Line: issue.Pos.Line - 1, Character: issue.Pos.Column - 1},
			},
			Severity: issue.DiagSeverity(),
			Source:   &issue.FromLinter,
			Message:  h.diagnosticMessage(&issue),
		}
		diagnostics = append(diagnostics, d)
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

		diagnostics, err := h.lint(uri)
		if err != nil {
			h.logger.Printf("%s", err)

			continue
		}

		if err := h.conn.Notify(
			context.Background(),
			"textDocument/publishDiagnostics",
			&PublishDiagnosticsParams{
				URI:         uri,
				Diagnostics: diagnostics,
			}); err != nil {
			h.logger.Printf("%s", err)
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
	case "workspace/didChangeConfiguration":
		return h.handlerWorkspaceDidChangeConfiguration(ctx, conn, req)
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
	h.command = params.InitializationOptions.Command

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

func (h *langHandler) handlerWorkspaceDidChangeConfiguration(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}
