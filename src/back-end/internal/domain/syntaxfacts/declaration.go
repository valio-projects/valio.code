package syntaxfacts

// Declaration is a written construct and its explicit source metadata.
type Declaration struct {
	ID             string      `json:"id"`             // ID is unique within a report.
	Name           string      `json:"name"`           // Name is written source spelling.
	Kind           string      `json:"kind"`           // Kind identifies class, struct, method or another declaration.
	ParentID       string      `json:"parentId"`       // ParentID links lexical containment.
	Start          int         `json:"start"`          // Start includes this UTF-8 byte offset.
	End            int         `json:"end"`            // End excludes this UTF-8 byte offset.
	Type           *string     `json:"type"`           // Type preserves an explicit type expression.
	UnderlyingType *string     `json:"underlyingType"` // UnderlyingType preserves explicit enum storage syntax.
	EnumValue      *string     `json:"enumValue"`      // EnumValue is a written expression, not constant evaluation.
	Visibility     string      `json:"visibility"`     // Visibility preserves explicit or unknown visibility.
	Modifiers      []string    `json:"modifiers"`      // Modifiers preserve written language flags.
	Attributes     []string    `json:"attributes"`     // Attributes are literal annotation syntax.
	Parameters     []Parameter `json:"parameters"`     // Parameters belong to this callable.
}
