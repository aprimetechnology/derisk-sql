package types

const (
	DiagnosticLevelFatal   = "FATAL"
	DiagnosticLevelWarning = "WARNING"
)

type Diagnostic struct {
	LineNumber   int    `json:"lineNumber"`
	LinePosition int    `json:"linePosition"`
	Text         string `json:"text"`
	Code         string `json:"code"`
	Level        string `json:"level"`
}

type Report struct {
	// Reference to the entire Migration object for context
	Migration   ParsedMigration   `json:"migration"`
	Text        string            `json:"text"`
	Diagnostics []Diagnostic      `json:"diagnostics,omitempty"`
	Actions     []string          `json:"actions"`
	Config      map[string]string `json:"config,omitempty"`
}
