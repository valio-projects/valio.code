package embeddings

// Completeness states whether an embedding record is safe to rank.
type Completeness string

const (
	// Ready means the record contains a complete finite vector and can be ranked.
	Ready Completeness = "ready"
	// Partial means the representation could not produce a rankable complete vector.
	Partial Completeness = "partial"
	// MissingProvider means no configured provider was available to create a vector.
	MissingProvider Completeness = "missing_provider"
)

// IsValid reports whether c is a recognized completeness state.
func (c Completeness) IsValid() bool {
	return c == Ready || c == Partial || c == MissingProvider
}
