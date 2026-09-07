package typeinfo

// TypeKind classifies a type form as reported by a language producer.
type TypeKind string

const (
	// TypeClass is an inheritable object type.
	TypeClass TypeKind = "class"
	// TypeStruct is a value or record-like structure.
	TypeStruct TypeKind = "struct"
	// TypeInterface defines a contract without concrete storage.
	TypeInterface TypeKind = "interface"
	// TypeEnum is a named set of typed constant values.
	TypeEnum TypeKind = "enum"
	// TypeAlias names another type without defining a new shape.
	TypeAlias TypeKind = "alias"
	// TypeArray has fixed-rank indexed elements.
	TypeArray TypeKind = "array"
	// TypeSlice is a dynamically sized sequence view.
	TypeSlice TypeKind = "slice"
	// TypeMap associates keys with values.
	TypeMap TypeKind = "map"
	// TypeFunction denotes a callable type.
	TypeFunction TypeKind = "function"
	// TypePointer denotes an addressable reference type.
	TypePointer TypeKind = "pointer"
	// TypeNamed denotes a named type whose form is producer-specific.
	TypeNamed TypeKind = "named"
	// TypePrimitive denotes a built-in scalar type.
	TypePrimitive TypeKind = "primitive"
	// TypeDelegate denotes a callable object type.
	TypeDelegate TypeKind = "delegate"
	// TypeRecord denotes a record-style aggregate.
	TypeRecord TypeKind = "record"
	// TypeUnknown preserves an unclassified producer result.
	TypeUnknown TypeKind = "unknown"
)
