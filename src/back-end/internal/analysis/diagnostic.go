package analysis

// Diagnostic records a parser or type-checker message and its optional source location.
type Diagnostic struct {
	// Message is the producer's human-readable diagnostic text.
	Message string `json:"message"`
	// Offset is the UTF-8 byte offset; zero may be unknown.
	Offset int `json:"offset"`
	// Line is the one-based source line; zero may be unknown.
	Line int `json:"line"`
	// Column is the one-based source column; zero may be unknown.
	Column int `json:"column"`
}
