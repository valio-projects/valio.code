package repositories

import (
	"context"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

// EmbeddingRepository stores immutable, workspace/view/profile-scoped vectors.
type EmbeddingRepository interface {
	Find(context.Context, domain.WorkspaceID, domain.ViewID, string, string, string) (embeddings.Record, bool, error)
	Save(context.Context, embeddings.Record) error
	List(context.Context, domain.WorkspaceID, domain.ViewID, string, int) ([]embeddings.Record, error)
}
