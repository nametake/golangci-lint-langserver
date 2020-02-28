package main

// InitializeResult is
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
}

// TextDocumentSyncKind is
type TextDocumentSyncKind int

// TDSKNone is
//nolint:unused,deadcode
const (
	TDSKNone TextDocumentSyncKind = iota
	TDSKFull
	TDSKIncremental
)

// CompletionProvider is
type CompletionProvider struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters"`
}

// ServerCapabilities is
type ServerCapabilities struct {
	TextDocumentSync           TextDocumentSyncKind `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionProvider  `json:"completionProvider,omitempty"`
	DocumentSymbolProvider     bool                 `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider         bool                 `json:"definitionProvider,omitempty"`
	DocumentFormattingProvider bool                 `json:"documentFormattingProvider,omitempty"`
	HoverProvider              bool                 `json:"hoverProvider,omitempty"`
	CodeActionProvider         bool                 `json:"codeActionProvider,omitempty"`
}
