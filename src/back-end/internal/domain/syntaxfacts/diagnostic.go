package syntaxfacts

// Diagnostic identifies an unsupported or invalid construct without copying source.
type Diagnostic struct {
	Kind  string `json:"kind"`           // Kind classifies the parser diagnostic.
	Code  string `json:"code,omitempty"` // Code is a stable machine-readable reason.
	Start int    `json:"start"`          // Start includes the diagnostic source offset.
	End   int    `json:"end"`            // End excludes the diagnostic source offset.
}
