package analysis

type Import struct {
	Path  string `json:"path"`
	Alias string `json:"alias,omitempty"`
	Range Range  `json:"range"`
}
