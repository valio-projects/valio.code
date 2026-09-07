// Package syntaxfacts defines the shared syntax-only interchange model.
package syntaxfacts

// Report describes written source constructs without compiler target resolution.
type Report struct {
	Schema       int           `json:"schema"`       // Schema pins the helper protocol.
	Language     string        `json:"language"`     // Language selects the source grammar.
	Path         string        `json:"path"`         // Path is repository-relative.
	ValidSyntax  bool          `json:"validSyntax"`  // ValidSyntax is false when parsing reports errors.
	Symbols      []Declaration `json:"symbols"`      // Symbols are declarations, not resolved identities.
	References   []Reference   `json:"references"`   // References retain unresolved written observations.
	Imports      []Import      `json:"imports"`      // Imports preserve source declarations.
	Diagnostics  []Diagnostic  `json:"diagnostics"`  // Diagnostics explain incomplete parsing.
	Capabilities Capabilities  `json:"capabilities"` // Capabilities describes this producer.
}
