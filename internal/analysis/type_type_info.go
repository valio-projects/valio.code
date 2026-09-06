package analysis

type TypeInfo struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	QualifiedName          string           `json:"qualifiedName"`
	Kind                   string           `json:"kind"`
	Alias                  bool             `json:"alias"`
	Exported               bool             `json:"exported"`
	Visibility             string           `json:"visibility"`
	UnderlyingType         TypeExpression   `json:"underlyingType"`
	ResolvedUnderlyingType string           `json:"resolvedUnderlyingType,omitempty"`
	TypeParameters         []ParameterInfo  `json:"typeParameters"`
	Fields                 []FieldInfo      `json:"fields"`
	Methods                []FunctionInfo   `json:"methods"`
	EmbeddedTypes          []TypeExpression `json:"embeddedTypes"`
	Constants              []ConstantInfo   `json:"constants"`
	Range                  Range            `json:"range"`
	NameRange              Range            `json:"nameRange"`
	Layout                 LayoutInfo       `json:"layout"`
	Evidence               string           `json:"evidence"`
}
