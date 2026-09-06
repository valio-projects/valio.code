package typeinfo

type ArrayDimension struct {
	Length     Fact[uint64] `json:"length"`
	LowerBound Fact[int64]  `json:"lowerBound"`
}
