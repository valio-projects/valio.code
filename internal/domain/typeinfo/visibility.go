package typeinfo

type Visibility string

const (
	VisibilityPublic            Visibility = "public"
	VisibilityPrivate           Visibility = "private"
	VisibilityProtected         Visibility = "protected"
	VisibilityInternal          Visibility = "internal"
	VisibilityProtectedInternal Visibility = "protected_internal" // family OR assembly (C#)
	VisibilityPrivateProtected  Visibility = "private_protected"  // family AND assembly (C#)
	VisibilityPackage           Visibility = "package"
	VisibilityModule            Visibility = "module"
	VisibilityUnknown           Visibility = "unknown"
)
