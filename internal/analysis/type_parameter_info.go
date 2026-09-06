package analysis

type ParameterInfo struct {
	Name     string         `json:"name"`
	Type     TypeExpression `json:"type"`
	Variadic bool           `json:"variadic"`
	Range    Range          `json:"range"`
}
