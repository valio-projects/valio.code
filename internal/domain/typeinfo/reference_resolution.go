package typeinfo

type ReferenceResolution string

const (
	ReferenceExact      ReferenceResolution = "exact"
	ReferenceCandidate  ReferenceResolution = "candidate"
	ReferenceUnresolved ReferenceResolution = "unresolved"
)
