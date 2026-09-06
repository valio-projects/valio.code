package typeinfo

type PropertyDescriptor struct {
	ID          string                `json:"id"`
	Name        Fact[string]          `json:"name"`
	Type        TypeReference         `json:"type"`
	Visibility  Fact[Visibility]      `json:"visibility"`
	Modifiers   Fact[[]Modifier]      `json:"modifiers"`
	Parameters  []ParameterDescriptor `json:"parameters"` // indexer parameters
	Getter      *MethodDescriptor     `json:"getter,omitempty"`
	Setter      *MethodDescriptor     `json:"setter,omitempty"`
	Attributes  []AttributeUse        `json:"attributes"`
	Declaration *OccurrenceLink       `json:"declaration,omitempty"`
}
