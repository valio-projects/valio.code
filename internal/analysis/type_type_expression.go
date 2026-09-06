package analysis

type TypeExpression struct {
	Text       string           `json:"text"`
	Kind       string           `json:"kind"`
	Element    *TypeExpression  `json:"element,omitempty"`
	Dimensions []ArrayDimension `json:"dimensions,omitempty"`
	Evidence   string           `json:"evidence"`
}
