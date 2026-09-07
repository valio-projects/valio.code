package queries

import (
	"context"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	"time"
)

// ProviderProbe contains no credential or source text; it verifies a configured
// model with a fixed synthetic input rather than merely checking an open port.
type ProviderProbe struct {
	ModelProfile string `json:"modelProfile"`
}

// ProbeProvider verifies reachability, vector validity and configured dimension.
// Returned model/revision fields are operator declarations, not runtime attestation.
func (s Service) ProbeProvider(ctx context.Context, q ProviderProbe) (map[string]any, error) {
	if s.Models == nil {
		return nil, fault.ErrUnavailable
	}
	profile, provider, err := s.Models.ResolveModel(q.ModelProfile, embeddings.Code)
	if err != nil || provider == nil {
		return nil, fault.ErrUnavailable
	}
	start := time.Now()
	vector, err := provider.Embed(ctx, profile, "valio.code embedding connectivity probe")
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fault.ErrUnavailable
	}
	if len(vector) != profile.Dimension {
		return nil, fault.ErrUnavailable
	}
	if _, err = embeddings.Cosine(vector, vector); err != nil {
		return nil, fault.ErrUnavailable
	}
	return map[string]any{"status": "ready", "dimension": len(vector), "configuredModel": profile.Model, "configuredRevision": profile.Revision, "profileFingerprint": profile.Fingerprint, "durationMs": time.Since(start).Milliseconds()}, nil
}
