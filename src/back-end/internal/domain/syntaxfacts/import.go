package syntaxfacts

// Import preserves a written import or using directive without loading dependencies.
type Import struct {
	Path       string `json:"path"`       // Path is the declared import target, not a local filesystem path.
	Kind       string `json:"kind"`       // Kind classifies the declaration.
	Resolution string `json:"resolution"` // Resolution must remain unresolved.
	Start      int    `json:"start"`      // Start is an inclusive UTF-8 byte offset.
	End        int    `json:"end"`        // End is an exclusive UTF-8 byte offset.
}
