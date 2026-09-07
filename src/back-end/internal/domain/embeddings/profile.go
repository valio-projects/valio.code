package embeddings

import (
	"errors"
	"strings"
)

// Profile pins a representation, model revision, distance metric, and vector dimension.
// Any change to these immutable vector-space fields produces a new fingerprint and requires reindexing.
type Profile struct {
	// ID is the stable caller-assigned profile identity.
	ID string `json:"id"`
	// Kind is the representation class accepted by this profile.
	Kind Kind `json:"kind"`
	// SchemaVersion identifies the template version that formed the representation.
	SchemaVersion int `json:"schemaVersion"`
	// Provider identifies the embedding service.
	Provider string `json:"provider"`
	// Model identifies the provider model.
	Model string `json:"model"`
	// Revision identifies the immutable provider-model revision or configured compatibility revision.
	Revision string `json:"revision"`
	// Distance is the vector comparison metric. Only cosine is currently supported.
	Distance string `json:"distance"`
	// Dimension is the exact number of float32 values emitted by the model.
	Dimension int `json:"dimension"`
	// Fingerprint is the canonical immutable identity derived from this profile's fields.
	Fingerprint string `json:"fingerprint"`
}

// NewProfile constructs a profile with its canonical immutable fingerprint.
func NewProfile(id string, kind Kind, schemaVersion int, provider, model, revision, distance string, dimension int) (Profile, error) {
	p := Profile{ID: id, Kind: kind, SchemaVersion: schemaVersion, Provider: provider, Model: model, Revision: revision, Distance: distance, Dimension: dimension}
	p.Fingerprint = FingerprintProfile(p)
	return p, p.Validate()
}

// Validate reports whether p is a canonical, self-consistent profile.
func (p Profile) Validate() error {
	if p.ID == "" || strings.TrimSpace(p.ID) != p.ID || !p.Kind.IsValid() || p.SchemaVersion < 1 ||
		p.Provider == "" || strings.TrimSpace(p.Provider) != p.Provider ||
		p.Model == "" || strings.TrimSpace(p.Model) != p.Model ||
		p.Revision == "" || strings.TrimSpace(p.Revision) != p.Revision ||
		p.Distance != "cosine" || p.Dimension < 1 || p.Dimension > 16384 ||
		p.Fingerprint != FingerprintProfile(p) {
		return errors.New("invalid embedding profile")
	}
	return nil
}
