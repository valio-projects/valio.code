package analysis

type ArrayDimension struct {
	Expression string `json:"expression"`
	Length     *int64 `json:"length,omitempty"`
	Resolution string `json:"resolution"`
}
