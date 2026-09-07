package search

// Range is a half-open UTF-8 byte range in a file's Content.
type Range struct {
	// Start is the included byte offset.
	Start int `json:"start"`
	// End is the excluded byte offset.
	End int `json:"end"`
}
