package main

type DocumentURI string

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

type TextDocumentItem struct {
	URI        DocumentURI `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidSaveTextDocumentParams struct {
	Text         *string                `json:"text"`
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               *string                        `json:"code,omitempty"`
	Source             *string                        `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}
