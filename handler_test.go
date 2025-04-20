package main

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func pt(s string) *string {
	return &s
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
				rootDir: filepath.Dir("./testdata/noconfig"),
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
				rootDir:      filepath.Dir("./testdata/nolintername"),
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
				rootDir: filepath.Dir("./testdata/loadconfig"),
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
				rootDir: filepath.Dir("./testdata/multifile"),
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
				rootDir: filepath.Dir("./testdata/nesteddir"),
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
				rootDir: filepath.Dir("./testdata/monorepo"),
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
				rootDir: filepath.Dir("./testdata/monorepo"),
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
