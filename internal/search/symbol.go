package search

type Symbol struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}
