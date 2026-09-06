package surreal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain/fault"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
)

type appEntry struct {
	// Table supplies the typed table value at this boundary.
	Table string `json:"table"`
	// Key supplies the typed key value at this boundary.
	Key string `json:"key"`
	// Payload supplies the typed payload value at this boundary.
	Payload any `json:"payload"`
}

// Publish atomically stores p after validating its expected head and immutable entries; ctx controls cancellation.
func (s *AppStore) Publish(ctx context.Context, p snapshots.Publication) error {
	if p.View.WorkspaceID != s.workspace || p.Snapshot.WorkspaceID != s.workspace {
		return fault.ErrForbidden
	}
	entries := []appEntry{{"source_snapshot", p.Snapshot.ID, p.Snapshot}}
	seen := map[string]bool{}
	for _, b := range p.Blobs {
		if !seen[b.ID] {
			entries = append(entries, appEntry{"source_blob", b.ID, b})
			seen[b.ID] = true
		}
	}
	// Membership belongs to the view. Source-file identity stores only immutable
	// repository/path/blob metadata; no mutable project membership is overwritten.
	for _, f := range p.View.Files {
		copy := f
		copy.ProjectIDs = nil
		entries = append(entries, appEntry{"source_file", f.ID, copy})
	}
	for _, a := range p.Artifacts {
		entries = append(entries, appEntry{"analysis_artifact", a.ID, a})
		for _, d := range a.Types {
			key := a.ID + "-" + string(d.Scope.ProjectID) + "-" + d.ID
			entries = append(entries, appEntry{"type_descriptor", key, d})
		}
	}
	for i, d := range p.View.Projects {
		entries = append(entries, appEntry{"project_revision", string(p.View.ProjectRevisions[i].ProjectRevisionID), d})
	}
	for i, hash := range p.Manifest.Shards {
		shardEntries := []any{}
		for _, e := range p.Manifest.Entries {
			sum := sha256.Sum256([]byte(e.Path))
			if int(sum[0]) == i {
				shardEntries = append(shardEntries, e)
			}
		}
		entries = append(entries, appEntry{"manifest_shard", p.Snapshot.ID + "-" + fmt.Sprintf("%02x", i), map[string]any{"snapshotId": p.Snapshot.ID, "index": i, "hash": hash, "entries": shardEntries}})
	}
	entries = append(entries, appEntry{"analysis_view", p.View.ID, p.View})
	// All staged immutable records and the publication pointer commit together.
	// A concurrent publisher/configuration change aborts the entire transaction.
	sql := `BEGIN TRANSACTION;
 LET $head = (SELECT VALUE payload FROM projection_head:⟨app-latest⟩);
 IF (array::len($head) = 0 AND $expected != '') OR (array::len($head) != 0 AND $head[0] != $expected) { THROW 'conflict'; };
 LET $currentDefinitions = (SELECT VALUE payload FROM project);
 IF array::len($currentDefinitions) != array::len($definitions) { THROW 'conflict'; };
 FOR $definition IN $definitions { LET $current = (SELECT VALUE payload FROM type::record('project',$definition.project.id)); IF array::len($current) != 1 OR $current[0] != $definition { THROW 'conflict'; }; };
 FOR $entry IN $entries {
  LET $existing = (SELECT VALUE payload FROM type::record($entry.table,$entry.key));
  IF array::len($existing) = 0 { CREATE type::record($entry.table,$entry.key) SET key=$entry.key,payload=$entry.payload; }
  ELSE IF $existing[0] != $entry.payload { THROW 'immutable conflict'; };
 };
 UPSERT projection_head:⟨app-latest⟩ SET key='app-latest',payload=$view;
 COMMIT TRANSACTION;`
	_, e := s.client.Query(ctx, sql, map[string]any{"entries": entries, "definitions": p.View.Projects, "expected": p.ExpectedHead, "view": p.View.ID})
	if e != nil {
		return fault.ErrConflict
	}
	return nil
}
func appDigest(values ...string) string {
	h := sha256.New()
	for _, s := range values {
		h.Write([]byte(s))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
