package analysis

type ConstantGroup struct {
	ID      string         `json:"id"`
	Kind    string         `json:"kind"`
	Range   Range          `json:"range"`
	Members []ConstantInfo `json:"members"`
}
