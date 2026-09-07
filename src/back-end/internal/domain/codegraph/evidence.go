package codegraph

// EvidenceOrigin identifies the producer class for a graph fact.
type EvidenceOrigin string

const (
	// OriginAnalyzer identifies a fact observed by this builder.
	OriginAnalyzer EvidenceOrigin = "analyzer"
)

// Evidence describes the producer and source span supporting a node or edge.
type Evidence struct {
	// Origin identifies the kind of producer that observed the fact.
	Origin EvidenceOrigin `json:"origin"`
	// Producer identifies the concrete parser or type checker profile.
	Producer string `json:"producer"`
	// Range identifies the supporting source range when a fact has one.
	Range *ByteRange `json:"range,omitempty"`
}
