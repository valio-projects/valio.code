package queries

// SyntaxQuery finds written declarations and members in a selected version.
// Names are syntactic lookup keys, never claims of compiler-resolved identity.
type SyntaxQuery struct {
	Scope  SearchScope `json:"scope"`
	FileID string      `json:"fileId,omitempty"`
	Name   string      `json:"name,omitempty"`
}
