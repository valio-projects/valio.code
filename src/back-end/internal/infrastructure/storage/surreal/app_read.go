package surreal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/analysis"
	"github.com/valio-projects/valio.code/internal/domain/fault"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/search"
	"strings"
)

func appMany[ // T supplies the typed t value at this boundary.
	T any](ctx context.Context, c *Client, table string, keys []string) ([]T, error) {
	result := []T{}
	for start := 0; start < len(keys); start += 200 {
		batch := keys[start:min(start+200, len(keys))]
		rows, e := c.Query(ctx, `LET $ids=$keys.map(|$key| type::record($table,$key)); SELECT VALUE payload FROM $ids;`, map[string]any{"table": table, "keys": batch})
		if e != nil {
			return nil, e
		}
		values, e := DecodeLast[[]T](rows)
		if e != nil {
			return nil, e
		}
		if len(values) != len(batch) {
			return nil, fault.ErrNotFound
		}
		result = append(result, values...)
	}
	return result, nil
}

// Files loads only source references in v and rejects exhausted scope budgets; ctx controls cancellation.
func (s *AppStore) Files(ctx context.Context, v snapshots.View) ([]search.File, error) {
	if v.WorkspaceID != s.workspace {
		return nil, fault.ErrForbidden
	}
	if len(v.Files) > 10000 {
		return nil, fault.ErrScopeTooLarge
	}
	total := 0
	keys := []string{}
	seen := map[string]bool{}
	for _, f := range v.Files {
		total += f.Size
		if total > 16<<20 {
			return nil, fault.ErrScopeTooLarge
		}
		if !seen[f.BlobID] {
			keys = append(keys, f.BlobID)
			seen[f.BlobID] = true
		}
	}
	blobs, e := appMany[snapshots.Blob](ctx, s.client, "source_blob", keys)
	if e != nil {
		return nil, e
	}
	byID := map[string]snapshots.Blob{}
	for _, b := range blobs {
		if b.WorkspaceID != s.workspace {
			return nil, fault.ErrForbidden
		}
		byID[b.ID] = b
	}
	artifacts, e := s.Artifacts(ctx, v)
	if e != nil {
		return nil, e
	}
	reports := map[string]analysis.Report{}
	syntax := map[string][]search.Symbol{}
	for _, a := range artifacts {
		mapped, err := syntaxSymbols(a.Syntax)
		if err != nil {
			return nil, fault.ErrInvalid
		}
		syntax[a.FileID] = mapped
		var r analysis.Report
		if json.Unmarshal(a.Report, &r) != nil {
			return nil, fault.ErrInvalid
		}
		reports[a.FileID] = r
	}
	result := make([]search.File, 0, len(v.Files))
	for _, f := range v.Files {
		b, ok := byID[f.BlobID]
		if !ok {
			return nil, fault.ErrNotFound
		}
		if len(b.Content) != f.Size {
			return nil, fault.ErrInvalid
		}
		file := search.File{ID: f.ID, Path: f.Path, Content: b.Content, ProjectIDs: f.ProjectIDs, RepositoryID: string(f.RepositoryID), SnapshotID: f.SnapshotID, Language: f.Language, Test: strings.HasSuffix(f.Path, "_test.go"), Generated: strings.Contains(b.Content, "Code generated")}
		for _, symbol := range reports[f.ID].Symbols {
			file.Symbols = append(file.Symbols, search.Symbol{Name: symbol.Name, Kind: symbol.Kind, Start: symbol.Range.Start, End: symbol.Range.End})
		}
		file.Symbols = append(file.Symbols, syntax[f.ID]...)
		result = append(result, file)
	}
	return result, nil
}

// Artifacts loads syntax/type evidence belonging only to v; ctx controls cancellation.
func (s *AppStore) Artifacts(ctx context.Context, v snapshots.View) ([]snapshots.Artifact, error) {
	if v.WorkspaceID != s.workspace {
		return nil, fault.ErrForbidden
	}
	keys := []string{}
	for _, f := range v.Files {
		if f.Language == "go" || v.Projections["retrieval_chunks"] == "ready" {
			b, _ := json.Marshal([]string{v.ID, f.ID})
			sum := sha256.Sum256(b)
			keys = append(keys, hex.EncodeToString(sum[:]))
		}
	}
	values, e := appMany[snapshots.Artifact](ctx, s.client, "analysis_artifact", keys)
	if e != nil {
		return nil, e
	}
	for _, a := range values {
		if a.ViewID != v.ID {
			return nil, fault.ErrForbidden
		}
	}
	return values, nil
}
