package analysis

type Symbol struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Kind             string `json:"kind"`
	Range            Range  `json:"range"`
	DeclarationRange Range  `json:"declarationRange"`
	Evidence         string `json:"evidence"`
}
