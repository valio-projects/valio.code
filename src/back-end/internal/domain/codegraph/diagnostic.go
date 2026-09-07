package codegraph

// Diagnostic is a bounded, source-safe build finding. It never stores source text.
type Diagnostic struct {
	// Code is a stable machine-readable category.
	Code string `json:"code"`
	// Message describes the issue without copying source content.
	Message string `json:"message"`
	// Range identifies the affected input source when available.
	Range *ByteRange `json:"range,omitempty"`
}
