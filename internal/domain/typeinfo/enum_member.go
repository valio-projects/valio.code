package typeinfo

type EnumMember struct {
	ID          string           `json:"id"`
	GroupID     string           `json:"groupId,omitempty"`
	Name        Fact[string]     `json:"name"`
	Expression  Fact[string]     `json:"expression"`
	Value       ConstantValue    `json:"value"`
	Attributes  []AttributeUse   `json:"attributes"`
	Declaration *OccurrenceLink  `json:"declaration,omitempty"`
	Occurrences []OccurrenceLink `json:"occurrences"`
}
