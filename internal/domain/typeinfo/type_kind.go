package typeinfo

type TypeKind string

const (
	TypeClass     TypeKind = "class"
	TypeStruct    TypeKind = "struct"
	TypeInterface TypeKind = "interface"
	TypeEnum      TypeKind = "enum"
	TypeAlias     TypeKind = "alias"
	TypeArray     TypeKind = "array"
	TypeSlice     TypeKind = "slice"
	TypeMap       TypeKind = "map"
	TypeFunction  TypeKind = "function"
	TypePointer   TypeKind = "pointer"
	TypeNamed     TypeKind = "named"
	TypePrimitive TypeKind = "primitive"
	TypeDelegate  TypeKind = "delegate"
	TypeRecord    TypeKind = "record"
	TypeUnknown   TypeKind = "unknown"
)
