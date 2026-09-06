package analysis

type Occurrence struct {
	Name       string `json:"name"`
	Range      Range  `json:"range"`
	Role       string `json:"role"`
	SymbolID   string `json:"symbolId,omitempty"`
	Resolution string `json:"resolution"`
	ObjectType string `json:"objectType,omitempty"`
}
