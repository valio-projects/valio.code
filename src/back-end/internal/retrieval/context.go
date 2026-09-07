package retrieval

// Context is deterministic, bounded source context with citations and omission
// explanations. ByteCount includes text, header, and citation identifiers.
type Context struct {
	// ViewID identifies the immutable view supplied by the caller.
	ViewID string `json:"viewId"`
	// Items contains selected source in deterministic relationship order.
	Items []ContextItem `json:"items"`
	// Omitted explains eligible relationship candidates not included in Items.
	Omitted []ContextOmission `json:"omitted"`
	// ByteCount is the exact byte count consumed from the configured budget.
	ByteCount int `json:"byteCount"`
}
