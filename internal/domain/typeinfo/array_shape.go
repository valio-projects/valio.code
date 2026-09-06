package typeinfo

// ArrayShape describes an array use, including rank and per-dimension bounds.
type ArrayShape struct {
	// ElementType is the type stored at each array position.
	ElementType TypeReference `json:"elementType"`
	// Rank is the number of dimensions when known.
	Rank Fact[int] `json:"rank"`
	// Dimensions has one item per rank when rank is known.
	Dimensions []ArrayDimension `json:"dimensions"`
}
