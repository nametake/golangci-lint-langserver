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

	testFilePath, _ := filepath.Abs("./testdata/simple/main.go")

	h := &langHandler{
		logger:  newStdLogger(false),
		command: []string{"golangci-lint", "run", "--out-format", "json", "--issues-exit-code=1"},
		rootDir: filepath.Dir(testFilePath),
	}

	testURI := DocumentURI("file://" + testFilePath)

	diagnostics, err := h.lint(testURI)
	if err != nil {
		t.Fatalf("lint() returned unexpected error: %v", err)
	}

	want := []Diagnostic{
		{
			Range: Range{
				Start: Position{
					Line:      2,
					Character: 4,
				},
				End: Position{
					Line:      2,
					Character: 4,
				},
			},
			Severity:           DSWarning,
			Code:               nil,
			Source:             pt("unused"),
			Message:            "unused: var `foo` is unused",
			RelatedInformation: nil,
		},
	}

	if diff := cmp.Diff(want, diagnostics); diff != "" {
		t.Errorf("lint() mismatch (-want +got):\n%s", diff)
	}
}
