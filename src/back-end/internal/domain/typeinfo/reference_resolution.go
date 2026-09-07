package typeinfo

// ReferenceResolution states the confidence of a symbol link.
type ReferenceResolution string

const (
	// ReferenceExact has one semantic target.
	ReferenceExact ReferenceResolution = "exact"
	// ReferenceCandidate retains one or more plausible targets.
	ReferenceCandidate ReferenceResolution = "candidate"
	// ReferenceUnresolved has no established target.
	ReferenceUnresolved ReferenceResolution = "unresolved"
)
