package typeinfo

// EnumDescriptor provides enum-specific members and their storage type.
type EnumDescriptor struct {
	// UnderlyingType is the type used to store each member value.
	UnderlyingType TypeReference `json:"underlyingType"`
	// Members contains the enum declarations in producer order.
	Members []EnumMember `json:"members"`
}
