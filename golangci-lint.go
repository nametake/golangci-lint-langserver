package main

import "strings"

// GolangCITextEdit represents a text edit from golangci-lint v2 SuggestedFixes
type GolangCITextEdit struct {
	Pos     int    `json:"Pos"`     // byte offset start
	End     int    `json:"End"`     // byte offset end
	NewText string `json:"NewText"` // base64 encoded replacement text
}

// SuggestedFix represents a suggested fix from golangci-lint v2
type SuggestedFix struct {
	Message   string             `json:"Message"`
	TextEdits []GolangCITextEdit `json:"TextEdits"`
}

type Issue struct {
	FromLinter     string         `json:"FromLinter"`
	Text           string         `json:"Text"`
	Severity       string         `json:"Severity"`
	SourceLines    []string       `json:"SourceLines"`
	SuggestedFixes []SuggestedFix `json:"SuggestedFixes"`
	Pos            struct {
		Filename string `json:"Filename"`
		Offset   int    `json:"Offset"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
	ExpectNoLint         bool   `json:"ExpectNoLint"`
	ExpectedNoLintLinter string `json:"ExpectedNoLintLinter"`
	LineRange            struct {
		From int `json:"From"`
		To   int `json:"To"`
	} `json:"LineRange,omitempty"`
}

func (i Issue) DiagSeverity() DiagnosticSeverity {
	if i.Severity == "" {
		// TODO: How to get default-severity from .golangci.yml, if available?
		i.Severity = defaultSeverity
	}

	switch strings.ToLower(i.Severity) {
	case "err", "error":
		return DSError
	case "warn", "warning":
		return DSWarning
	case "info", "information":
		return DSInformation
	case "hint":
		return DSHint
	default:
		return DSWarning
	}
}

type GolangCILintResult struct {
	Issues []Issue `json:"Issues"`
	Report struct {
		Linters []struct {
			Name             string `json:"Name"`
			Enabled          bool   `json:"Enabled"`
			EnabledByDefault bool   `json:"EnabledByDefault,omitempty"`
		} `json:"Linters"`
		Error string `json:"Error"`
	} `json:"Report"`
}
