package main

type InitializeParams struct {
	RootURI string `json:"rootUri,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
}

type TextDocumentSyncKind int

//nolint:unused,deadcode
const (
	TDSKNone TextDocumentSyncKind = iota
	TDSKFull
	TDSKIncremental
)

type CompletionProvider struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters"`
}

type ServerCapabilities struct {
	TextDocumentSync           TextDocumentSyncKind `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionProvider  `json:"completionProvider,omitempty"`
	DocumentSymbolProvider     bool                 `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider         bool                 `json:"definitionProvider,omitempty"`
	DocumentFormattingProvider bool                 `json:"documentFormattingProvider,omitempty"`
	HoverProvider              bool                 `json:"hoverProvider,omitempty"`
	CodeActionProvider         bool                 `json:"codeActionProvider,omitempty"`
}
