package typeinfo

type LayoutKind string

const (
	LayoutAuto       LayoutKind = "auto"
	LayoutSequential LayoutKind = "sequential"
	LayoutExplicit   LayoutKind = "explicit"
	LayoutUnknown    LayoutKind = "unknown"
)
