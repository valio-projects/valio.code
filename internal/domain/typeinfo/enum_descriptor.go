package typeinfo

type EnumDescriptor struct {
	UnderlyingType TypeReference `json:"underlyingType"`
	Members        []EnumMember  `json:"members"`
}
