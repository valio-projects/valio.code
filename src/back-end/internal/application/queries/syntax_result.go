package queries

import "encoding/json"

// SyntaxResult carries selected multi-language source reports with their files.
type SyntaxResult struct {
	ViewID  string       `json:"viewId"`
	Reports []SyntaxFile `json:"reports"`
	Status  string       `json:"status"`
}

// SyntaxFile retains source coordinates and language-specific written metadata.
type SyntaxFile struct {
	FileID       string          `json:"fileId"`
	RepositoryID string          `json:"repositoryId"`
	Path         string          `json:"path"`
	Language     string          `json:"language"`
	Report       json.RawMessage `json:"report"`
}
