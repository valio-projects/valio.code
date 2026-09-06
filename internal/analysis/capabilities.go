package analysis

type Capabilities struct {
	SyntaxOutline      bool `json:"syntaxOutline"`
	DefinitionRanges   bool `json:"definitionRanges"`
	TypeChecked        bool `json:"typeChecked"`
	ResolvedReferences bool `json:"resolvedReferences"`
	ResolvedCalls      bool `json:"resolvedCalls"`
	ControlFlow        bool `json:"controlFlow"`
}
