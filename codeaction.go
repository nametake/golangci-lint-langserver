package main

import (
	"encoding/base64"
	"encoding/json"
)

// convertSuggestedFixes converts golangci-lint v2 SuggestedFixes to LSP TextEdits
func (h *langHandler) convertSuggestedFixes(fixes []SuggestedFix, fileContent []byte) *DiagnosticFixData {
	var edits []TextEdit

	for _, fix := range fixes {
		for _, te := range fix.TextEdits {
			// Decode base64 NewText
			newText, err := base64.StdEncoding.DecodeString(te.NewText)
			if err != nil {
				h.logger.Printf("Failed to decode base64 NewText: %v", err)
				continue
			}

			// Convert byte offsets to line/character positions
			startPos := h.offsetToPosition(fileContent, te.Pos)
			endOffset := te.End

			// When deleting entire line(s) from column 0, include the trailing newline
			// to avoid leaving empty lines
			if len(newText) == 0 && startPos.Character == 0 &&
				endOffset < len(fileContent) && fileContent[endOffset] == '\n' {
				endOffset++
			}

			endPos := h.offsetToPosition(fileContent, endOffset)

			edit := TextEdit{
				Range: Range{
					Start: startPos,
					End:   endPos,
				},
				NewText: string(newText),
			}
			edits = append(edits, edit)
		}
	}

	if len(edits) == 0 {
		return nil
	}

	return &DiagnosticFixData{Edits: edits}
}

// offsetToPosition converts a byte offset to an LSP Position (line/character)
func (h *langHandler) offsetToPosition(content []byte, offset int) Position {
	if offset < 0 {
		return Position{Line: 0, Character: 0}
	}
	if offset > len(content) {
		offset = len(content)
	}

	line := 0
	lineStart := 0

	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	character := offset - lineStart
	return Position{Line: line, Character: character}
}

// extractWorkspaceEdit extracts workspace edit from diagnostic data
func (h *langHandler) extractWorkspaceEdit(uri DocumentURI, diag Diagnostic) *WorkspaceEdit {
	if diag.Data == nil {
		return nil
	}

	// Data comes back from JSON as map[string]any, need to re-marshal and unmarshal
	b, err := json.Marshal(diag.Data)
	if err != nil {
		return nil
	}

	var fixData DiagnosticFixData
	if err := json.Unmarshal(b, &fixData); err == nil && len(fixData.Edits) > 0 {
		return &WorkspaceEdit{
			Changes: map[DocumentURI][]TextEdit{
				uri: fixData.Edits,
			},
		}
	}

	return nil
}
