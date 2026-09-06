package analysis

type ConstantInfo struct {
	ID                  string       `json:"id"`
	GroupID             string       `json:"groupId"`
	Name                string       `json:"name"`
	Exported            bool         `json:"exported"`
	DeclaredType        string       `json:"declaredType,omitempty"`
	ResolvedType        string       `json:"resolvedType,omitempty"`
	Expression          string       `json:"expression"`
	InheritedExpression bool         `json:"inheritedExpression"`
	Value               string       `json:"value,omitempty"`
	Resolution          string       `json:"resolution"`
	Range               Range        `json:"range"`
	References          []Occurrence `json:"references"`
}
