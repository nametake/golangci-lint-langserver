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
	lintPackage  bool

	// published tracks, per package directory, the documents (path -> URI)
	// that currently have non-empty diagnostics published, so they can be
	// cleared once their issues are fixed. Only accessed from the linter
	// goroutine.
	published map[string]map[string]DocumentURI

	rootURI string
	rootDir string
}

// As defined in the `golangci-lint` source code:
// https://github.com/golangci/golangci-lint/blob/main/pkg/exitcodes/exitcodes.go#L24
const GoNoFilesExitCode = 5

func findModuleRoot(filePath, fallback string) string {
	directory := filepath.Dir(filePath)
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return fallback
}

func (h *langHandler) errToDiagnostics(err error) []Diagnostic {
	var message string
	switch e := err.(type) {
	case *exec.ExitError:
		if e.ExitCode() == GoNoFilesExitCode {
			return []Diagnostic{}
		}
		message = string(e.Stderr)
	default:
		h.logger.DebugJSON("golangci-lint-langserver: errToDiagnostics message", message)
		message = e.Error()
	}
	return []Diagnostic{
		{Severity: DSError, Message: message},
	}
}

// lint returns the diagnostics for the given document only.
func (h *langHandler) lint(uri DocumentURI) ([]Diagnostic, error) {
	files, err := h.lintFiles(uri)
	if err != nil {
		return nil, err
	}

	return files[uri], nil
}

// lintFiles runs golangci-lint on the package (directory) of the given
// document. The result always contains an entry for uri. If lintPackage is
// enabled, it also contains entries for the other files of the package that
// have issues.
func (h *langHandler) lintFiles(uri DocumentURI) (map[DocumentURI][]Diagnostic, error) {
	diagnostics := make([]Diagnostic, 0)
	files := map[DocumentURI][]Diagnostic{uri: diagnostics}

	path := uriToPath(string(uri))
	dir, file := filepath.Split(path)

	moduleRoot := findModuleRoot(path, h.rootDir)

	args := make([]string, 0, len(h.command))
	args = append(args, h.command[1:]...)
	args = append(args, dir)
	cmd := exec.Command(h.command[0], args...)
	if strings.HasPrefix(path, moduleRoot) {
		cmd.Dir = moduleRoot
		file = path[len(moduleRoot)+1:]
	} else {
		cmd.Dir = dir
	}

	h.logger.DebugJSON("golangci-lint-langserver: golingci-lint cmd:", cmd.Args)

	b, err := cmd.Output()
	if err == nil {
		return files, nil
	} else if len(b) == 0 {
		// golangci-lint would output critical error to stderr rather than stdout
		// https://github.com/nametake/golangci-lint-langserver/issues/24
		files[uri] = h.errToDiagnostics(err)

		return files, nil
	}

	var result GolangCILintResult
	if err := json.Unmarshal(b, &result); err != nil {
		files[uri] = h.errToDiagnostics(err)

		return files, nil
	}

	h.logger.DebugJSON("golangci-lint-langserver: result:", result)

	pkgDir := filepath.Clean(dir)

	for _, issue := range result.Issues {
		issueURI := uri
		if file != issue.Pos.Filename {
			if !h.lintPackage {
				continue
			}

			// Issue filenames are relative to the directory golangci-lint was run in.
			issuePath := issue.Pos.Filename
			if !filepath.IsAbs(issuePath) {
				issuePath = filepath.Join(cmd.Dir, issuePath)
			}

			// Only report files of the linted package, not of other directories.
			if filepath.Dir(issuePath) != pkgDir {
				continue
			}

			issueURI = pathToURI(issuePath)
		}

		d := Diagnostic{
			Range: Range{
				Start: Position{
					Line:      max(issue.Pos.Line-1, 0),
					Character: max(issue.Pos.Column-1, 0),
				},
				End: Position{
					Line:      max(issue.Pos.Line-1, 0),
					Character: max(issue.Pos.Column-1, 0),
				},
			},
			Severity: issue.DiagSeverity(),
			Source:   &issue.FromLinter,
			Message:  h.diagnosticMessage(&issue),
		}
		files[issueURI] = append(files[issueURI], d)
	}

	return files, nil
}

// publishParams builds the publishDiagnostics notifications for a lint run of
// the package containing uri. Documents of that package which had diagnostics
// published by a previous run but have none now are cleared.
func (h *langHandler) publishParams(uri DocumentURI, files map[DocumentURI][]Diagnostic) []PublishDiagnosticsParams {
	params := make([]PublishDiagnosticsParams, 0, len(files))

	if h.lintPackage {
		if h.published == nil {
			h.published = make(map[string]map[string]DocumentURI)
		}

		pkgDir := filepath.Dir(uriToPath(string(uri)))

		// Files are compared by path, as the client and pathToURI may encode
		// the URI of the same file differently.
		current := make(map[string]DocumentURI, len(files))
		for u, diagnostics := range files {
			if len(diagnostics) > 0 {
				current[uriToPath(string(u))] = u
			}
		}

		inResult := make(map[string]struct{}, len(files))
		for u := range files {
			inResult[uriToPath(string(u))] = struct{}{}
		}

		// Clear stale diagnostics first, so they never override fresh ones.
		for path, u := range h.published[pkgDir] {
			if _, ok := inResult[path]; !ok {
				params = append(params, PublishDiagnosticsParams{URI: u, Diagnostics: []Diagnostic{}})
			}
		}

		h.published[pkgDir] = current
	}

	for u, diagnostics := range files {
		params = append(params, PublishDiagnosticsParams{URI: u, Diagnostics: diagnostics})
	}

	return params
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

		files, err := h.lintFiles(uri)
		if err != nil {
			h.logger.Printf("%s\n", err)

			continue
		}

		for _, params := range h.publishParams(uri, files) {
			if err := h.conn.Notify(
				context.Background(),
				"textDocument/publishDiagnostics",
				&params,
			); err != nil {
				h.logger.Printf("%s\n", err)
			}
		}
	}
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
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

func (h *langHandler) handleInitialize(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.rootURI = params.RootURI
	h.rootDir = uriToPath(params.RootURI)
	h.conn = conn
	h.command = params.InitializationOptions.Command
	h.lintPackage = params.InitializationOptions.LintPackage

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

func (h *langHandler) handleShutdown(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result any, err error) {
	close(h.request)

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidOpen(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidClose(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result any, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidChange(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result any, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidSave(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}

func (h *langHandler) handlerWorkspaceDidChangeConfiguration(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result any, err error) {
	return nil, nil
}
