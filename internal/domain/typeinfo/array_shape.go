package typeinfo

type ArrayShape struct {
	ElementType TypeReference    `json:"elementType"`
	Rank        Fact[int]        `json:"rank"`
	Dimensions  []ArrayDimension `json:"dimensions"`
}
