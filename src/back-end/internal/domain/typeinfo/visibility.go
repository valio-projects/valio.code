package typeinfo

// Visibility records a language-level access domain.
type Visibility string

const (
	// VisibilityPublic permits access from any consumer.
	VisibilityPublic Visibility = "public"
	// VisibilityPrivate limits access to the declaring type.
	VisibilityPrivate Visibility = "private"
	// VisibilityProtected permits derived-type access.
	VisibilityProtected Visibility = "protected"
	// VisibilityInternal limits access to an assembly or module boundary.
	VisibilityInternal Visibility = "internal"
	// VisibilityProtectedInternal permits family or assembly access in C#.
	VisibilityProtectedInternal Visibility = "protected_internal"
	// VisibilityPrivateProtected requires family and assembly access in C#.
	VisibilityPrivateProtected Visibility = "private_protected"
	// VisibilityPackage limits access to a package.
	VisibilityPackage Visibility = "package"
	// VisibilityModule limits access to a module.
	VisibilityModule Visibility = "module"
	// VisibilityUnknown preserves an unavailable access classification.
	VisibilityUnknown Visibility = "unknown"
)
