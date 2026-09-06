package typeinfo

// Modifier is extensible so a producer can preserve language-specific flags.
type Modifier string

const (
	// ModifierVirtual permits dispatch through an override.
	ModifierVirtual Modifier = "virtual"
	// ModifierOverride replaces an inherited implementation.
	ModifierOverride Modifier = "override"
	// ModifierAbstract requires an implementation elsewhere.
	ModifierAbstract Modifier = "abstract"
	// ModifierFinal prohibits further overriding or extension.
	ModifierFinal Modifier = "final"
	// ModifierSealed prevents inheritance according to the language rules.
	ModifierSealed Modifier = "sealed"
	// ModifierStatic belongs to a type rather than an instance.
	ModifierStatic Modifier = "static"
	// ModifierReadonly restricts writes after initialization.
	ModifierReadonly Modifier = "readonly"
	// ModifierConst identifies a compile-time constant declaration.
	ModifierConst Modifier = "const"
	// ModifierEmbedded promotes an embedded member or type.
	ModifierEmbedded Modifier = "embedded"
	// ModifierVariadic accepts a variable number of final arguments.
	ModifierVariadic Modifier = "variadic"
)
