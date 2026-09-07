package embeddings

import (
	"testing"
	"time"

	"github.com/valio-projects/valio.code/internal/domain"
)

func TestProfileFingerprintIsCanonicalAndChangesWithImmutableFields(t *testing.T) {
	profile, err := NewProfile("code-v1", Code, 1, "provider", "model", "revision", "cosine", 2)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Fingerprint != FingerprintProfile(profile) {
		t.Fatal("profile fingerprint depends on itself")
	}
	mutations := []Profile{
		func() Profile { value := profile; value.ID = "other"; return value }(),
		func() Profile { value := profile; value.Kind = Symbol; return value }(),
		func() Profile { value := profile; value.SchemaVersion++; return value }(),
		func() Profile { value := profile; value.Provider = "other"; return value }(),
		func() Profile { value := profile; value.Model = "other"; return value }(),
		func() Profile { value := profile; value.Revision = "other"; return value }(),
		func() Profile { value := profile; value.Dimension++; return value }(),
	}
	for _, mutation := range mutations {
		if FingerprintProfile(mutation) == profile.Fingerprint {
			t.Fatalf("immutable profile mutation did not change fingerprint: %+v", mutation)
		}
		if mutation.Validate() == nil {
			t.Fatal("mutated profile retained an old fingerprint")
		}
	}
}

func TestSupportedKindsAndIncompleteRecords(t *testing.T) {
	kinds := []Kind{Code, Symbol, Context, Documentation, Architecture, Change, Error, API}
	for _, kind := range kinds {
		if !kind.IsValid() {
			t.Fatalf("unsupported planned representation kind %q", kind)
		}
	}
	if Kind("unknown").IsValid() {
		t.Fatal("accepted unknown representation kind")
	}
	profile, _ := NewProfile("code-v1", Code, 1, "provider", "model", "revision", "cosine", 2)
	input := Input{ID: "input", WorkspaceID: domain.WorkspaceID("workspace"), ViewID: domain.ViewID("view"), Kind: Code, Representation: "approved", PolicyApproved: true}
	input.Fingerprint = Hash(input.Representation)
	record, err := NewRecord(profile, input, []float32{1, 2}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	record.Completeness = Partial
	if err = record.Validate(); err == nil {
		t.Fatal("partial record carried a rankable vector")
	}
	record = Record{ID: record.ID, WorkspaceID: record.WorkspaceID, ViewID: record.ViewID, InputID: record.InputID, Kind: record.Kind, ProfileFingerprint: record.ProfileFingerprint, Provider: record.Provider, Model: record.Model, Revision: record.Revision, Distance: record.Distance, Dimension: record.Dimension, InputFingerprint: record.InputFingerprint, Completeness: MissingProvider, CreatedAt: record.CreatedAt}
	if err = record.Validate(); err != nil {
		t.Fatal(err)
	}
	record.ID = "changed"
	if err = record.Validate(); err == nil {
		t.Fatal("record identity no longer matched its immutable scope")
	}
}
