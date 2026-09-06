package typeinfo

// LayoutKind identifies how a target language lays out a type.
type LayoutKind string

const (
	// LayoutAuto leaves placement to the runtime or compiler.
	LayoutAuto LayoutKind = "auto"
	// LayoutSequential follows declaration order subject to target rules.
	LayoutSequential LayoutKind = "sequential"
	// LayoutExplicit uses producer-established member offsets.
	LayoutExplicit LayoutKind = "explicit"
	// LayoutUnknown indicates no supported layout evidence exists.
	LayoutUnknown LayoutKind = "unknown"
)
