package typeinfo

// ArrayDimension records one dimension of an array type.
type ArrayDimension struct {
	// Length is the element count for this dimension when the producer knows it.
	Length Fact[uint64] `json:"length"`
	// LowerBound is the first valid index; it is not assumed to be zero.
	LowerBound Fact[int64] `json:"lowerBound"`
}
