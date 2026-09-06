package types

// MemberCounts counts recorded records, not the total possible semantic members.
// In particular zero occurrences means no recorded usages, not proven unused.
type MemberCounts struct {
	Fields            int `json:"fields"`
	Properties        int `json:"properties"`
	Methods           int `json:"methods"`
	Constructors      int `json:"constructors"`
	Parameters        int `json:"parameters"`
	Returns           int `json:"returns"`
	GenericParameters int `json:"genericParameters"`
	Attributes        int `json:"attributes"`
	EnumMembers       int `json:"enumMembers"`
	Constants         int `json:"constants"`
	Occurrences       int `json:"occurrences"`
}
