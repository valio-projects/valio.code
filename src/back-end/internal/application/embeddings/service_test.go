package embeddings

import (
	"context"
	"testing"
	"time"

	"github.com/valio-projects/valio.code/internal/domain"
	domainembeddings "github.com/valio-projects/valio.code/internal/domain/embeddings"
)

type memoryRepo struct {
	value domainembeddings.Record
	found bool
}

func (m *memoryRepo) Find(context.Context, domain.WorkspaceID, domain.ViewID, string, string, string) (domainembeddings.Record, bool, error) {
	return m.value, m.found, nil
}
func (m *memoryRepo) Save(_ context.Context, v domainembeddings.Record) error {
	m.value = v
	m.found = true
	return nil
}
func (m *memoryRepo) List(context.Context, domain.WorkspaceID, domain.ViewID, string, int) ([]domainembeddings.Record, error) {
	return nil, nil
}

type calls struct{ n int }

func (c *calls) Embed(context.Context, domainembeddings.Profile, string) ([]float32, error) {
	c.n++
	return []float32{1, 2}, nil
}
func TestComputeExistingDoesNotCallProvider(t *testing.T) {
	p := domainembeddings.Profile{ID: "p", Kind: domainembeddings.Code, SchemaVersion: 1, Provider: "x", Model: "m", Revision: "r", Distance: "cosine", Dimension: 2}
	p.Fingerprint = domainembeddings.FingerprintProfile(p)
	in := domainembeddings.Input{ID: "i", WorkspaceID: "w", ViewID: "v", Kind: domainembeddings.Code, Fingerprint: domainembeddings.Hash("approved"), Representation: "approved", PolicyApproved: true}
	existing, _ := domainembeddings.NewRecord(p, in, []float32{1, 2}, time.Now())
	repo := &memoryRepo{value: existing, found: true}
	provider := &calls{}
	got, err := (Service{Repository: repo, Provider: provider}).Compute(context.Background(), p, in)
	if err != nil || got.ID != existing.ID || provider.n != 0 {
		t.Fatal("existing record executed provider")
	}
}
