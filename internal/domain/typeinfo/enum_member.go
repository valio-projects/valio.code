package typeinfo

// EnumMember describes a typed constant or enum declaration.
type EnumMember struct {
	// ID is stable within the containing type descriptor.
	ID string `json:"id"`
	// GroupID identifies the source declaration group when applicable.
	GroupID string `json:"groupId,omitempty"`
	// Name is the declared member name.
	Name Fact[string] `json:"name"`
	// Expression is the source expression, which may be inherited or unknown.
	Expression Fact[string] `json:"expression"`
	// Value retains the typed constant value when resolved.
	Value ConstantValue `json:"value"`
	// Attributes lists annotations on the member.
	Attributes []AttributeUse `json:"attributes"`
	// Declaration identifies the defining source occurrence when known.
	Declaration *OccurrenceLink `json:"declaration,omitempty"`
	// Occurrences lists source references to this member.
	Occurrences []OccurrenceLink `json:"occurrences"`
}
