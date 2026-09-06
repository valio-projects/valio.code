package typeinfo

type AttributeValueKind string

const (
	AttributeLiteral    AttributeValueKind = "literal"
	AttributeExpression AttributeValueKind = "expression"
	AttributeRedacted   AttributeValueKind = "redacted"
)
