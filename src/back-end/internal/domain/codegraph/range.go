package codegraph

// ByteRange is a half-open UTF-8 byte range in one supplied file.
type ByteRange struct {
	// FileID identifies the supplied file containing Start and End.
	FileID string `json:"fileId"`
	// Start is the included byte offset in the source content.
	Start int `json:"start"`
	// End is the excluded byte offset in the source content.
	End int `json:"end"`
}
