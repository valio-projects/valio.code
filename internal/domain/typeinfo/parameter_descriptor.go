package typeinfo

type ParameterDescriptor struct {
	ID           string           `json:"id"`
	Name         Fact[string]     `json:"name"`
	Position     Fact[int]        `json:"position"`
	Type         TypeReference    `json:"type"`
	Modifiers    Fact[[]Modifier] `json:"modifiers"`
	DefaultValue Fact[string]     `json:"defaultValue"`
	Attributes   []AttributeUse   `json:"attributes"`
}
