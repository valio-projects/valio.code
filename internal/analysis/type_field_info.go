package analysis

type FieldInfo struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       TypeExpression `json:"type"`
	Exported   bool           `json:"exported"`
	Visibility string         `json:"visibility"`
	Embedded   bool           `json:"embedded"`
	Tag        string         `json:"tag,omitempty"`
	TagLiteral string         `json:"tagLiteral,omitempty"`
	Range      Range          `json:"range"`
	NameRange  Range          `json:"nameRange"`
}
