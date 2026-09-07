package analysis

// Import records one syntax-level import declaration.
type Import struct {
	// Path is the unquoted import path.
	Path string `json:"path"`
	// Alias is empty when no explicit alias was written.
	Alias string `json:"alias,omitempty"`
	// Range covers the full import declaration.
	Range Range `json:"range"`
}
