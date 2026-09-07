package search

// Symbol is optional searchable declaration metadata for a file.
type Symbol struct {
	// Name is the producer-recorded symbol name.
	Name string `json:"name"`
	// Kind is the producer-recorded symbol category.
	Kind string `json:"kind"`
	// Start is the included source byte offset.
	Start int `json:"start"`
	// End is the excluded source byte offset.
	End int `json:"end"`
}
