package retrieval

// RepresentationKind identifies one independently versioned retrieval view.
// The enum describes available profiles; a chunk carries only profiles backed by
// facts present in its source and parsed metadata.
type RepresentationKind string

const (
	// RepresentationCode preserves canonical source code.
	RepresentationCode RepresentationKind = "code"
	// RepresentationSymbol records a declaration name and signature.
	RepresentationSymbol RepresentationKind = "symbol"
	// RepresentationDocumentation records an attached source comment.
	RepresentationDocumentation RepresentationKind = "documentation"
	// RepresentationContext records generated local scope metadata.
	RepresentationContext RepresentationKind = "context"
	// RepresentationArchitecture is reserved for explicit architecture facts.
	RepresentationArchitecture RepresentationKind = "architecture"
	// RepresentationChange is reserved for explicit change-history facts.
	RepresentationChange RepresentationKind = "change"
	// RepresentationError records detected source-level error facts.
	RepresentationError RepresentationKind = "error"
	// RepresentationAPI records an exported API declaration.
	RepresentationAPI RepresentationKind = "api"
)
