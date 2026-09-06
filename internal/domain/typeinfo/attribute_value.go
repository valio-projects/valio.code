package typeinfo

// AttributeValue preserves typed literals as canonical text, avoiding JSON
// float precision loss. Expression text is not claimed to be evaluated. Redacted
// values have no Literal or Expression payload.
type AttributeValue struct {
	Kind            AttributeValueKind `json:"kind"`
	Type            TypeReference      `json:"type"`
	Literal         Fact[string]       `json:"literal"`
	Expression      Fact[string]       `json:"expression"`
	RedactionReason string             `json:"redactionReason,omitempty"`
}
