package embeddings

import (
	"errors"
	"strings"
	"time"

	"github.com/valio-projects/valio.code/internal/domain"
)

// Record is an immutable vector result scoped to one workspace, view, input, and profile.
// Only Ready records carry a vector.
type Record struct {
	// ID is the deterministic identity of workspace, view, input, profile, and input fingerprints.
	ID string `json:"id"`
	// WorkspaceID scopes the record to one workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// ViewID scopes the record to one immutable view.
	ViewID domain.ViewID `json:"viewId"`
	// InputID identifies the caller's input within WorkspaceID and ViewID.
	InputID string `json:"inputId"`
	// Kind identifies the representation class used for this vector.
	Kind Kind `json:"kind"`
	// ProfileFingerprint identifies the immutable model vector space.
	ProfileFingerprint string `json:"profileFingerprint"`
	// Provider identifies the service that produced or was expected to produce the vector.
	Provider string `json:"provider"`
	// Model identifies the configured model.
	Model string `json:"model"`
	// Revision identifies the configured model revision.
	Revision string `json:"revision"`
	// Distance identifies the supported comparison metric.
	Distance string `json:"distance"`
	// Dimension is the exact vector length required by the profile.
	Dimension int `json:"dimension"`
	// InputFingerprint binds the record to the transient representation without retaining that text.
	InputFingerprint string `json:"inputFingerprint"`
	// Completeness tells readers whether Vector is rankable.
	Completeness Completeness `json:"completeness"`
	// Vector is a finite, nonzero vector only when Completeness is Ready.
	Vector []float32 `json:"vector,omitempty"`
	// CreatedAt is the UTC instant at which this immutable record was made.
	CreatedAt time.Time `json:"createdAt"`
}

// Validate reports whether r is canonical, immutable, and internally consistent.
func (r Record) Validate() error {
	if r.ID == "" || r.WorkspaceID == "" || r.ViewID == "" || r.InputID == "" || strings.TrimSpace(r.InputID) != r.InputID ||
		!r.Kind.IsValid() || !fingerprint(r.ProfileFingerprint) || !fingerprint(r.InputFingerprint) ||
		r.Provider == "" || strings.TrimSpace(r.Provider) != r.Provider ||
		r.Model == "" || strings.TrimSpace(r.Model) != r.Model ||
		r.Revision == "" || strings.TrimSpace(r.Revision) != r.Revision ||
		r.Distance != "cosine" || r.Dimension < 1 || r.Dimension > 16384 || r.CreatedAt.IsZero() ||
		r.ID != recordID(r.WorkspaceID, r.ViewID, r.InputID, r.ProfileFingerprint, r.InputFingerprint) {
		return errors.New("invalid embedding record")
	}
	if !r.Completeness.IsValid() {
		return errors.New("invalid embedding completeness")
	}
	if r.Completeness != Ready {
		if len(r.Vector) != 0 {
			return errors.New("incomplete embedding record carries a vector")
		}
		return nil
	}
	if len(r.Vector) != r.Dimension || !finiteNonZero(r.Vector) {
		return errors.New("invalid embedding vector")
	}
	return nil
}

// NewRecord binds one provider response to validated profile and input values.
func NewRecord(profile Profile, input Input, vector []float32, now time.Time) (Record, error) {
	if err := profile.Validate(); err != nil {
		return Record{}, err
	}
	if err := input.Validate(); err != nil {
		return Record{}, err
	}
	if profile.Kind != input.Kind || len(vector) != profile.Dimension || !finiteNonZero(vector) {
		return Record{}, errors.New("embedding model response does not match profile")
	}
	r := Record{ID: recordID(input.WorkspaceID, input.ViewID, input.ID, profile.Fingerprint, input.Fingerprint), WorkspaceID: input.WorkspaceID, ViewID: input.ViewID, InputID: input.ID, Kind: input.Kind, ProfileFingerprint: profile.Fingerprint, Provider: profile.Provider, Model: profile.Model, Revision: profile.Revision, Distance: profile.Distance, Dimension: profile.Dimension, InputFingerprint: input.Fingerprint, Completeness: Ready, Vector: append([]float32(nil), vector...), CreatedAt: now.UTC()}
	return r, r.Validate()
}

// MissingRecord reports an unavailable configured provider without fabricating a vector.
func MissingRecord(profile Profile, input Input, now time.Time) (Record, error) {
	if err := profile.Validate(); err != nil {
		return Record{}, err
	}
	if err := input.Validate(); err != nil || profile.Kind != input.Kind {
		return Record{}, errors.New("invalid missing-provider input")
	}
	r := Record{ID: recordID(input.WorkspaceID, input.ViewID, input.ID, profile.Fingerprint, input.Fingerprint), WorkspaceID: input.WorkspaceID, ViewID: input.ViewID, InputID: input.ID, Kind: input.Kind, ProfileFingerprint: profile.Fingerprint, Provider: profile.Provider, Model: profile.Model, Revision: profile.Revision, Distance: profile.Distance, Dimension: profile.Dimension, InputFingerprint: input.Fingerprint, Completeness: MissingProvider, CreatedAt: now.UTC()}
	return r, r.Validate()
}

func recordID(workspace domain.WorkspaceID, view domain.ViewID, inputID, profileFingerprint, inputFingerprint string) string {
	return Hash(string(workspace), string(view), inputID, profileFingerprint, inputFingerprint)
}
