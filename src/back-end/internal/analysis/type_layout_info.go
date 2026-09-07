package analysis

type LayoutInfo struct {
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	SizeBytes      *int64 `json:"sizeBytes,omitempty"`
	AlignmentBytes *int64 `json:"alignmentBytes,omitempty"`
}
