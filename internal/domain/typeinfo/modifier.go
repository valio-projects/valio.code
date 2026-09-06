package typeinfo

// Modifier is extensible so a producer can preserve language-specific flags.
type Modifier string

const (
	ModifierVirtual  Modifier = "virtual"
	ModifierOverride Modifier = "override"
	ModifierAbstract Modifier = "abstract"
	ModifierFinal    Modifier = "final"
	ModifierSealed   Modifier = "sealed"
	ModifierStatic   Modifier = "static"
	ModifierReadonly Modifier = "readonly"
	ModifierConst    Modifier = "const"
	ModifierEmbedded Modifier = "embedded"
	ModifierVariadic Modifier = "variadic"
)
