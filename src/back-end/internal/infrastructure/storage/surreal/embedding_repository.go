package surreal

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

// EmbeddingRepository persists flexible immutable payloads in embedding_record.
// The table requires an explicit migration only when schemafull constraints/indexes are introduced.
type EmbeddingRepository struct {
	client  *Client
	records *Repository[embeddings.Record]
}

// NewEmbeddingRepository constructs the immutable SurrealDB adapter for embedding records.
func NewEmbeddingRepository(client *Client) (*EmbeddingRepository, error) {
	r, e := NewRepository[embeddings.Record](client, "embedding_record", true)
	if e != nil {
		return nil, e
	}
	return &EmbeddingRepository{client: client, records: r}, nil
}

// Find returns one record only when all deterministic workspace, view, input, and fingerprint scope values match.
func (r *EmbeddingRepository) Find(ctx context.Context, workspace domain.WorkspaceID, view domain.ViewID, inputID, profile, input string) (embeddings.Record, bool, error) {
	if workspace == "" || view == "" || inputID == "" || !embeddingFingerprint(profile) || !embeddingFingerprint(input) {
		return embeddings.Record{}, false, errors.New("invalid embedding scope")
	}
	value, e := r.records.Find(ctx, embeddings.Hash(string(workspace), string(view), inputID, profile, input))
	if errors.Is(e, ErrNotFound) {
		return embeddings.Record{}, false, nil
	}
	if e != nil {
		return embeddings.Record{}, false, e
	}
	if value.WorkspaceID != workspace || value.ViewID != view || value.InputID != inputID || value.ProfileFingerprint != profile || value.InputFingerprint != input {
		return embeddings.Record{}, false, errors.New("embedding scope mismatch")
	}
	if e = value.Validate(); e != nil {
		return embeddings.Record{}, false, errors.New("invalid persisted embedding record")
	}
	return value, true, nil
}

// Save creates value once. A second byte-for-byte identical value succeeds, while any changed value with the same identity fails.
func (r *EmbeddingRepository) Save(ctx context.Context, value embeddings.Record) error {
	if e := value.Validate(); e != nil {
		return e
	}
	_, e := r.client.Query(ctx, `BEGIN TRANSACTION;
LET $existing=(SELECT VALUE payload FROM type::record($table,$key));
IF array::len($existing)=0 { CREATE type::record($table,$key) SET key=$key,payload=$payload; }
ELSE IF $existing[0]=$payload { } ELSE { THROW 'embedding identity conflict'; };
COMMIT TRANSACTION;`, map[string]any{"table": "embedding_record", "key": value.ID, "payload": value})
	return e
}

// List returns at most limit records after applying workspace, view, and profile filters.
// It rejects an oversized scoped result so callers must narrow their request rather than silently losing candidates.
func (r *EmbeddingRepository) List(ctx context.Context, workspace domain.WorkspaceID, view domain.ViewID, profile string, limit int) ([]embeddings.Record, error) {
	if workspace == "" || view == "" || !embeddingFingerprint(profile) || limit < 1 || limit > 1000 {
		return nil, errors.New("invalid embedding list scope")
	}
	rows, e := r.client.Query(ctx, `SELECT VALUE payload FROM type::table($table) WHERE payload.workspaceId=$workspace AND payload.viewId=$view AND payload.profileFingerprint=$profile ORDER BY key LIMIT $limit;`, map[string]any{"table": "embedding_record", "workspace": workspace, "view": view, "profile": profile, "limit": limit + 1})
	if e != nil {
		return nil, e
	}
	values, e := DecodeLast[[]embeddings.Record](rows)
	if e != nil {
		return nil, e
	}
	if len(values) > limit {
		return nil, errors.New("embedding list truncated; refine scope")
	}
	for _, value := range values {
		if e := value.Validate(); e != nil {
			return nil, e
		}
	}
	return values, nil
}

func embeddingFingerprint(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
