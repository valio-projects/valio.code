package typeinfo

// TypeDescriptor is an evidence-backed description of one type in one scope.
type TypeDescriptor struct {
	// ID is the producer's stable type identity.
	ID string `json:"id"`
	// Scope pins all facts to workspace, project, profile, and analysis version.
	Scope TypeScope `json:"scope"`
	// Name is the declared type name.
	Name Fact[string] `json:"name"`
	// FullyQualifiedName is the rendered qualified name when known.
	FullyQualifiedName Fact[string] `json:"fullyQualifiedName"`
	// Language identifies the source language when collected.
	Language Fact[string] `json:"language"`
	// Kind classifies the type form.
	Kind Fact[TypeKind] `json:"kind"`
	// Visibility records access scope.
	Visibility Fact[Visibility] `json:"visibility"`
	// Modifiers preserves language-level type flags.
	Modifiers Fact[[]Modifier] `json:"modifiers"`
	// Symbol carries exact, candidate, or unresolved semantic linkage.
	Symbol SymbolReference `json:"symbol"`
	// Declaration identifies the defining occurrence when known.
	Declaration *OccurrenceLink `json:"declaration,omitempty"`
	// Fields lists declared fields.
	Fields []FieldDescriptor `json:"fields"`
	// Properties lists declared properties.
	Properties []PropertyDescriptor `json:"properties"`
	// Methods lists non-constructor callable members.
	Methods []MethodDescriptor `json:"methods"`
	// Constructors lists construction callables where the language has them.
	Constructors []MethodDescriptor `json:"constructors"`
	// GenericParameters lists type parameters.
	GenericParameters []GenericParameter `json:"genericParameters"`
	// Attributes lists type-level annotations.
	Attributes []AttributeUse `json:"attributes"`
	// UnderlyingType identifies the represented type for aliases or similar forms.
	UnderlyingType *TypeReference `json:"underlyingType,omitempty"`
	// Enum contains enum-specific metadata when applicable.
	Enum *EnumDescriptor `json:"enum,omitempty"`
	// Constants lists typed constant members without asserting enum semantics.
	Constants []EnumMember `json:"constants"`
	// EmbeddedTypes lists promoted or embedded type uses.
	EmbeddedTypes []TypeReference `json:"embeddedTypes"`
	// Array contains array-shape metadata when this type is array-shaped.
	Array *ArrayShape `json:"array,omitempty"`
	// Layout contains target-specific physical facts or their unknown reasons.
	Layout TypeLayout `json:"layout"`
}
