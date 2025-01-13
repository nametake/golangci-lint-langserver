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
	_, err := exec.LookPath("golangci-lint")
	if err != nil {
		t.Fatal("golangci-lint is not installed in this environment")
	}

	tests := []struct {
		name     string
		h        *langHandler
		filePath string
		want     []Diagnostic
	}{
		{
			name: "simple",
			h: &langHandler{
				logger:  newStdLogger(false),
				command: []string{"golangci-lint", "run", "--out-format", "json", "--issues-exit-code=1"},
				rootDir: filepath.Dir("./testdata/simple"),
			},
			filePath: "./testdata/simple/main.go",
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
					Message:            "unused: var `foo` is unused",
					RelatedInformation: nil,
				},
			},
		},
		{
			name: "nolintername",
			h: &langHandler{
				logger:       newStdLogger(false),
				command:      []string{"golangci-lint", "run", "--out-format", "json", "--issues-exit-code=1"},
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
					Message:            "var `foo` is unused",
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
