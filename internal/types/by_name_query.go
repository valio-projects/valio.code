package types

type ByNameQuery struct {
	Scope QueryScope `json:"scope"`
	Name  string     `json:"name"`
}
