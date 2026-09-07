package analysis

type FunctionInfo struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Signature      string          `json:"signature"`
	Exported       bool            `json:"exported"`
	Visibility     string          `json:"visibility"`
	Receiver       *ParameterInfo  `json:"receiver,omitempty"`
	TypeParameters []ParameterInfo `json:"typeParameters"`
	Parameters     []ParameterInfo `json:"parameters"`
	Returns        []ParameterInfo `json:"returns"`
	Range          Range           `json:"range"`
	NameRange      Range           `json:"nameRange"`
	Evidence       string          `json:"evidence"`
}
