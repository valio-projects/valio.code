package analysis

// Occurrence records one identifier use or declaration with resolution evidence.
type Occurrence struct {
	// Name is the source identifier text.
	Name string `json:"name"`
	// Range is the identifier's half-open UTF-8 byte span.
	Range Range `json:"range"`
	// Role distinguishes definition, use, blank, and package-clause occurrences.
	Role string `json:"role"`
	// SymbolID identifies the resolved or syntax-defined symbol when available.
	SymbolID string `json:"symbolId,omitempty"`
	// Resolution states how SymbolID was established or why it is unavailable.
	Resolution string `json:"resolution"`
	// ObjectType names the Go semantic object type when collected.
	ObjectType string `json:"objectType,omitempty"`
}
