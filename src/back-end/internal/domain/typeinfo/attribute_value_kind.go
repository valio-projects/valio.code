package typeinfo

// AttributeValueKind identifies how an attribute argument was retained.
type AttributeValueKind string

const (
	// AttributeLiteral holds canonical literal text.
	AttributeLiteral AttributeValueKind = "literal"
	// AttributeExpression holds unevaluated expression text.
	AttributeExpression AttributeValueKind = "expression"
	// AttributeRedacted omits the value and supplies a redaction reason.
	AttributeRedacted AttributeValueKind = "redacted"
)
