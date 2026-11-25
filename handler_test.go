package main

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func unusedDiagnostic(name string) Diagnostic {
	return Diagnostic{
		Range: Range{
			Start: Position{Line: 3, Character: 4},
			End:   Position{Line: 3, Character: 4},
		},
		Severity: DSWarning,
		Source:   pt("unused"),
		Message:  "unused: var " + name + " is unused",
	}
}

func TestLangHandler_lintFiles_Integration(t *testing.T) {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Fatal("golangci-lint is not installed in this environment")
	}

	command := []string{"golangci-lint", "run", "--output.json.path", "stdout", "--issues-exit-code=1", "--show-stats=false"}

	mainURI := DocumentURI("file://" + mustAbs(t, "./testdata/lintpackage/main.go"))
	barURI := pathToURI(mustAbs(t, "./testdata/lintpackage/bar.go"))

	tests := []struct {
		name        string
		lintPackage bool
		want        map[DocumentURI][]Diagnostic
	}{
		{
			name:        "lintPackage disabled: only the linted file",
			lintPackage: false,
			want: map[DocumentURI][]Diagnostic{
				mainURI: {unusedDiagnostic("foo")},
			},
		},
		{
			name:        "lintPackage enabled: all files of the package, not sub packages",
			lintPackage: true,
			want: map[DocumentURI][]Diagnostic{
				mainURI: {unusedDiagnostic("foo")},
				barURI:  {unusedDiagnostic("bar")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &langHandler{
				logger:      newStdLogger(false),
				command:     command,
				rootDir:     mustAbs(t, "./testdata/lintpackage"),
				lintPackage: tt.lintPackage,
			}

			files, err := h.lintFiles(mainURI)
			if err != nil {
				t.Fatalf("lintFiles() returned unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, files); diff != "" {
				t.Errorf("lintFiles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLangHandler_publishParams(t *testing.T) {
	sortParams := cmpopts.SortSlices(func(a, b PublishDiagnosticsParams) bool { return a.URI < b.URI })

	mainURI := DocumentURI("file:///project/pkg/main.go")
	barURI := DocumentURI("file:///project/pkg/bar.go")
	otherURI := DocumentURI("file:///project/other/other.go")

	t.Run("lintPackage disabled publishes the result as is", func(t *testing.T) {
		h := &langHandler{}

		got := h.publishParams(mainURI, map[DocumentURI][]Diagnostic{mainURI: {}})
		want := []PublishDiagnosticsParams{{URI: mainURI, Diagnostics: []Diagnostic{}}}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("publishParams() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("lintPackage enabled clears fixed files of the same package only", func(t *testing.T) {
		h := &langHandler{lintPackage: true}

		// First run: bar.go has an issue.
		h.publishParams(mainURI, map[DocumentURI][]Diagnostic{
			mainURI: {},
			barURI:  {unusedDiagnostic("bar")},
		})
		// Other package has an issue too.
		h.publishParams(otherURI, map[DocumentURI][]Diagnostic{
			otherURI: {unusedDiagnostic("other")},
		})

		// Second run: bar.go got fixed.
		got := h.publishParams(mainURI, map[DocumentURI][]Diagnostic{mainURI: {}})
		want := []PublishDiagnosticsParams{
			{URI: barURI, Diagnostics: []Diagnostic{}},
			{URI: mainURI, Diagnostics: []Diagnostic{}},
		}

		if diff := cmp.Diff(want, got, sortParams); diff != "" {
			t.Errorf("publishParams() mismatch (-want +got):\n%s", diff)
		}

		// Third run: nothing left to clear.
		got = h.publishParams(mainURI, map[DocumentURI][]Diagnostic{mainURI: {}})
		want = []PublishDiagnosticsParams{{URI: mainURI, Diagnostics: []Diagnostic{}}}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("publishParams() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("lintPackage enabled matches files by path, not URI encoding", func(t *testing.T) {
		h := &langHandler{lintPackage: true}

		encoded := DocumentURI("file:///project/my%20pkg/bar.go")
		decoded := DocumentURI("file:///project/my pkg/bar.go")
		main := DocumentURI("file:///project/my%20pkg/main.go")

		h.publishParams(main, map[DocumentURI][]Diagnostic{encoded: {unusedDiagnostic("bar")}})

		// Same file reported under a differently encoded URI must not be cleared.
		got := h.publishParams(main, map[DocumentURI][]Diagnostic{decoded: {unusedDiagnostic("bar")}})
		want := []PublishDiagnosticsParams{{URI: decoded, Diagnostics: []Diagnostic{unusedDiagnostic("bar")}}}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("publishParams() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestPathToURI(t *testing.T) {
	path := mustAbs(t, "./testdata/my dir/main.go")

	uri := pathToURI(path)
	if got := uriToPath(string(uri)); got != path {
		t.Errorf("uriToPath(pathToURI(%q)) = %q", path, got)
	}
}
