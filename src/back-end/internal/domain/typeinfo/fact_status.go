package typeinfo

// FactStatus states whether a fact value has usable evidence.
type FactStatus string

const (
	// FactKnown has a value, an explicit owner scope, and evidence.
	FactKnown FactStatus = "known"
	// FactUnresolved was expected but could not be established.
	FactUnresolved FactStatus = "unresolved"
	// FactUnsupported cannot be collected by the active producer.
	FactUnsupported FactStatus = "unsupported"
)
