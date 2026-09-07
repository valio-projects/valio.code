package syntaxfacts

// Parameter describes a written argument declaration without type resolution.
type Parameter struct {
	Name       string   `json:"name"`       // Name is the written binding spelling.
	Type       *string  `json:"type"`       // Type is an explicit type expression or unknown.
	Start      int      `json:"start"`      // Start is an inclusive UTF-8 byte offset.
	End        int      `json:"end"`        // End is an exclusive UTF-8 byte offset.
	Modifiers  []string `json:"modifiers"`  // Modifiers includes ref/out and other written flags.
	Attributes []string `json:"attributes"` // Attributes preserves literal annotation lists.
}
