package structuregraph

// ByteRange is a half-open UTF-8 byte range in one supplied source file.
type ByteRange struct {
	// FileID identifies the file that contains Start through End.
	FileID string `json:"fileId"`
	// Start is the included UTF-8 byte offset.
	Start int `json:"start"`
	// End is the excluded UTF-8 byte offset.
	End int `json:"end"`
}
