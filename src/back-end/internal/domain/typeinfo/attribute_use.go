package typeinfo

// AttributeUse records an annotation applied to a declaration or type use.
type AttributeUse struct {
	// ID distinguishes this application from other uses of the same class.
	ID string `json:"id"`
	// RawSyntax retains a complete annotation list when the producer cannot
	// safely split its class name or arguments. It is source evidence, not an
	// evaluated attribute argument.
	RawSyntax *Fact[string] `json:"rawSyntax,omitempty"`
	// Name is the source-level attribute name when available.
	Name Fact[string] `json:"name"`
	// AttributeClass resolves the annotation's declaring symbol when possible.
	AttributeClass SymbolReference `json:"attributeClass"`
	// PositionalArguments preserves arguments in source order.
	PositionalArguments []AttributeValue `json:"positionalArguments"`
	// NamedArguments preserves explicitly named arguments.
	NamedArguments []NamedAttributeArgument `json:"namedArguments"`
	// Occurrence links this use to source evidence when a span is available.
	Occurrence *OccurrenceLink `json:"occurrence,omitempty"`
}
