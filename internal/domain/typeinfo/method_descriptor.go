package typeinfo

type MethodDescriptor struct {
	ID                string                `json:"id"`
	Name              Fact[string]          `json:"name"`
	Signature         Fact[string]          `json:"signature"`
	Visibility        Fact[Visibility]      `json:"visibility"`
	Modifiers         Fact[[]Modifier]      `json:"modifiers"`
	Receiver          *ParameterDescriptor  `json:"receiver,omitempty"`
	Parameters        []ParameterDescriptor `json:"parameters"`
	Returns           []ReturnDescriptor    `json:"returns"`
	GenericParameters []GenericParameter    `json:"genericParameters"`
	Attributes        []AttributeUse        `json:"attributes"`
	Declaration       *OccurrenceLink       `json:"declaration,omitempty"`
}
