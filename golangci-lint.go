package main

//nolint:unused,deadcode
type GolangCILintResult struct {
	Issues []struct {
		FromLinter  string      `json:"FromLinter"`
		Text        string      `json:"Text"`
		SourceLines []string    `json:"SourceLines"`
		Replacement interface{} `json:"Replacement"`
		Pos         struct {
			Filename string `json:"Filename"`
			Offset   int    `json:"Offset"`
			Line     int    `json:"Line"`
			Column   int    `json:"Column"`
		} `json:"Pos"`
		LineRange struct {
			From int `json:"From"`
			To   int `json:"To"`
		} `json:"LineRange,omitempty"`
	} `json:"Issues"`
	Report struct {
		Linters []struct {
			Name             string `json:"Name"`
			Enabled          bool   `json:"Enabled"`
			EnabledByDefault bool   `json:"EnabledByDefault,omitempty"`
		} `json:"Linters"`
	} `json:"Report"`
}
