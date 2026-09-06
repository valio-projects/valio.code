package typeinfo

type TypeDescriptor struct {
	ID                 string               `json:"id"`
	Scope              TypeScope            `json:"scope"`
	Name               Fact[string]         `json:"name"`
	FullyQualifiedName Fact[string]         `json:"fullyQualifiedName"`
	Language           Fact[string]         `json:"language"`
	Kind               Fact[TypeKind]       `json:"kind"`
	Visibility         Fact[Visibility]     `json:"visibility"`
	Modifiers          Fact[[]Modifier]     `json:"modifiers"`
	Symbol             SymbolReference      `json:"symbol"`
	Declaration        *OccurrenceLink      `json:"declaration,omitempty"`
	Fields             []FieldDescriptor    `json:"fields"`
	Properties         []PropertyDescriptor `json:"properties"`
	Methods            []MethodDescriptor   `json:"methods"`
	Constructors       []MethodDescriptor   `json:"constructors"`
	GenericParameters  []GenericParameter   `json:"genericParameters"`
	Attributes         []AttributeUse       `json:"attributes"`
	UnderlyingType     *TypeReference       `json:"underlyingType,omitempty"`
	Enum               *EnumDescriptor      `json:"enum,omitempty"`
	Constants          []EnumMember         `json:"constants"` // typed constant members; not a claim that the language has enums
	EmbeddedTypes      []TypeReference      `json:"embeddedTypes"`
	Array              *ArrayShape          `json:"array,omitempty"`
	Layout             TypeLayout           `json:"layout"`
}
