package typeinfo

type AttributeUse struct {
	ID                  string                   `json:"id"`
	Name                Fact[string]             `json:"name"`
	AttributeClass      SymbolReference          `json:"attributeClass"`
	PositionalArguments []AttributeValue         `json:"positionalArguments"`
	NamedArguments      []NamedAttributeArgument `json:"namedArguments"`
	Occurrence          *OccurrenceLink          `json:"occurrence,omitempty"`
}
