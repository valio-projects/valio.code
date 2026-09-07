package analysis

// Capabilities states which analysis facts were actually collected; false means unavailable.
type Capabilities struct {
	// SyntaxOutline reports syntax-level declaration collection.
	SyntaxOutline bool `json:"syntaxOutline"`
	// DefinitionRanges reports source definition spans.
	DefinitionRanges bool `json:"definitionRanges"`
	// TypeChecked reports successful local type checking.
	TypeChecked bool `json:"typeChecked"`
	// ResolvedReferences reports semantic reference resolution.
	ResolvedReferences bool `json:"resolvedReferences"`
	// ResolvedCalls reports semantic call resolution.
	ResolvedCalls bool `json:"resolvedCalls"`
	// ControlFlow reports control-flow extraction.
	ControlFlow bool `json:"controlFlow"`
}
