package typeinfo

type FactStatus string

const (
	FactKnown       FactStatus = "known"
	FactUnresolved  FactStatus = "unresolved"
	FactUnsupported FactStatus = "unsupported"
)
