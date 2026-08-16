package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/jsonrpc2"
)

func pt(s string) *string {
	return &s
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestLangHandlerPublishesFreshDiagnosticsAfterWatchedFileChanges(t *testing.T) {
	// Enable the subprocess-only behavior in TestLintCommandHelper.
	t.Setenv("GO_WANT_LINT_HELPER", "1")

	dir := t.TempDir()
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/stale\n\ngo 1.23\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("package main\n\n// lint-error\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	serverStream, clientStream := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	diagnostics := make(chan PublishDiagnosticsParams, 2)
	clientHandler := jsonrpc2.HandlerWithError(func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		if req.Method != "textDocument/publishDiagnostics" {
			return nil, nil
		}
		var params PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		diagnostics <- params
		return nil, nil
	})

	serverConn := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(serverStream, jsonrpc2.VSCodeObjectCodec{}), NewHandler(newStdLogger(false), false))
	clientConn := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(clientStream, jsonrpc2.VSCodeObjectCodec{}), clientHandler)
	defer serverConn.Close()
	defer clientConn.Close()

	var initialized InitializeResult
	if err := clientConn.Call(ctx, "initialize", InitializeParams{
		RootURI: "file://" + dir,
		InitializationOptions: InitializationOptions{
			Command: []string{os.Args[0], "-test.run=TestLintCommandHelper", "--"},
		},
	}, &initialized); err != nil {
		t.Fatal(err)
	}
	if initialized.Capabilities.TextDocumentSync.Change != TDSKNone {
		t.Fatalf("text document sync kind = %d, want none", initialized.Capabilities.TextDocumentSync.Change)
	}

	uri := DocumentURI("file://" + filePath)
	if err := clientConn.Notify(ctx, "textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, LanguageID: "go", Version: 0},
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-diagnostics:
		if len(got.Diagnostics) != 1 {
			t.Fatalf("initial diagnostics count = %d, want 1", len(got.Diagnostics))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial diagnostics")
	}

	updated := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(filePath, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.Notify(ctx, "workspace/didChangeWatchedFiles", map[string]any{
		"changes": []map[string]any{{"uri": uri, "type": 2}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.Notify(ctx, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 1},
		"contentChanges": []map[string]any{{"text": updated}},
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-diagnostics:
		if len(got.Diagnostics) != 0 {
			t.Fatalf("updated diagnostics count = %d, want 0", len(got.Diagnostics))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for updated diagnostics")
	}
}

func TestLintCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_LINT_HELPER") != "1" {
		return
	}

	dir := os.Args[len(os.Args)-1]
	contents, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}
	if !strings.Contains(string(contents), "lint-error") {
		os.Exit(0)
	}

	fmt.Fprint(os.Stdout, `{"Issues":[{"FromLinter":"test","Text":"stale diagnostic","Pos":{"Filename":"main.go","Line":3,"Column":1}}]}`)
	os.Exit(1)
}

func TestLangHandler_lint_Integration(t *testing.T) {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Fatal("golangci-lint is not installed in this environment")
	}

	command := []string{"golangci-lint", "run", "--output.json.path", "stdout", "--issues-exit-code=1", "--show-stats=false"}

	tests := []struct {
		name     string
		h        *langHandler
		filePath string
		want     []Diagnostic
	}{
		{
			name: "no config file",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/noconfig"),
			},
			filePath: "./testdata/noconfig/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "unused: var foo is unused",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "nolintername option works as expected",
			h: &langHandler{
				logger:       newStdLogger(false),
				command:      command,
				rootDir:      mustAbs(t, "./testdata/nolintername"),
				noLinterName: true,
			},
			filePath: "./testdata/nolintername/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "var foo is unused",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "config file is loaded successfully",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/loadconfig"),
			},
			filePath: "./testdata/loadconfig/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      8,
							Character: 0,
						},
						End: Position{
							Line:      8,
							Character: 0,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("wsl"),
					Message:            "wsl: block should not end with a whitespace (or comment)",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "multiple files in rootDir",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/multifile"),
			},
			filePath: "./testdata/multifile/bar.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "unused: var bar is unused",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "nested directories in rootDir",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/nesteddir"),
			},
			filePath: "./testdata/nesteddir/bar/bar.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "unused: var bar is unused",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "monorepo with multiple go.mod and .golangci.yaml files (foo module)",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/monorepo"),
			},
			filePath: "./testdata/monorepo/foo/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      8,
							Character: 0,
						},
						End: Position{
							Line:      8,
							Character: 0,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("wsl"),
					Message:            "wsl: block should not end with a whitespace (or comment)",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "monorepo with multiple go.mod and .golangci.yaml files (bar module)",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/monorepo"),
			},
			filePath: "./testdata/monorepo/bar/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "unused: var foo is unused",
					RelatedInformation: nil,
				},
				{
					Range: Range{
						Start: Position{
							Line:      8,
							Character: 0,
						},
						End: Position{
							Line:      8,
							Character: 0,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("wsl"),
					Message:            "wsl: block should not end with a whitespace (or comment)",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "nested go.mod: file in root module uses root .golangci.yaml (wsl only)",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/nestedmod"),
			},
			filePath: "./testdata/nestedmod/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      8,
							Character: 0,
						},
						End: Position{
							Line:      8,
							Character: 0,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("wsl"),
					Message:            "wsl: block should not end with a whitespace (or comment)",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "nested go.mod: file in sub module uses sub .golangci.yaml (unused only)",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: command,
				rootDir: mustAbs(t, "./testdata/nestedmod"),
			},
			filePath: "./testdata/nestedmod/sub/main.go",
			want: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      3,
							Character: 4,
						},
						End: Position{
							Line:      3,
							Character: 4,
						},
					},
					Severity:           DSWarning,
					Code:               nil,
					Source:             pt("unused"),
					Message:            "unused: var foo is unused",
					RelatedInformation: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFilePath, err := filepath.Abs(tt.filePath)
			if err != nil {
				t.Fatalf("filepath.Abs() returned unexpected error: %v", err)
			}
			testURI := DocumentURI("file://" + testFilePath)
			diagnostics, err := tt.h.lint(testURI)
			if err != nil {
				t.Fatalf("lint() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, diagnostics); diff != "" {
				t.Errorf("lint() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
