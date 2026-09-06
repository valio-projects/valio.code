package typeinfo

type FieldDescriptor struct {
	ID          string           `json:"id"`
	Name        Fact[string]     `json:"name"`
	Type        TypeReference    `json:"type"`
	Visibility  Fact[Visibility] `json:"visibility"`
	Modifiers   Fact[[]Modifier] `json:"modifiers"`
	Attributes  []AttributeUse   `json:"attributes"`
	Tag         Fact[string]     `json:"tag"` // Go struct tag; not a CLR attribute.
	Declaration *OccurrenceLink  `json:"declaration,omitempty"`
}
