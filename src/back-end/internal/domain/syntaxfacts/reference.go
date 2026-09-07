package syntaxfacts

// Reference records written identifier or call-target spelling, without a binding.
type Reference struct {
	Name       string `json:"name"`       // Name is observed spelling.
	Kind       string `json:"kind"`       // Kind distinguishes identifier and call observations.
	Resolution string `json:"resolution"` // Resolution must remain unresolved.
	Start      int    `json:"start"`      // Start is an inclusive UTF-8 byte offset.
	End        int    `json:"end"`        // End is an exclusive UTF-8 byte offset.
}
