package snapshots

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain/retrieval"
	chunking "github.com/valio-projects/valio.code/internal/retrieval"
)

// buildChunks reads only the already-sanitized immutable content map and groups
// deterministic chunks by their source-file identity for atomic publication.
func buildChunks(ctx context.Context, v View, contents map[string]string) (map[string][]retrieval.Chunk, error) {
	sources := make([]chunking.Source, 0, len(v.Files))
	for _, f := range v.Files {
		sources = append(sources, chunking.Source{ID: f.ID, RepositoryID: string(f.RepositoryID), Path: f.Path, Language: f.Language, Content: contents[f.ID], ProjectIDs: f.ProjectIDs})
	}
	chunks, err := (chunking.Builder{}).Build(ctx, sources)
	if err != nil {
		return nil, err
	}
	byFile := map[string][]retrieval.Chunk{}
	for _, chunk := range chunks {
		byFile[chunk.FileID] = append(byFile[chunk.FileID], chunk)
	}
	return byFile, nil
}
