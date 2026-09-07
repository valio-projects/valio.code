package syntaxfacts

// Capabilities separates supported syntax from unsupported compiler semantics.
type Capabilities struct {
	Syntax             bool `json:"syntax"`             // Syntax declares availability of the grammar.
	SemanticResolution bool `json:"semanticResolution"` // SemanticResolution must be false for this protocol.
}
