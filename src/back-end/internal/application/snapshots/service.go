package snapshots

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/analysis"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
	"github.com/valio-projects/valio.code/internal/projects"
	"github.com/valio-projects/valio.code/internal/types"
	"sort"
)

// MaxFiles bounds the total number of source records in one published view.
const MaxFiles = 10000

// MaxBytes bounds both incoming JSON and combined source bytes.
const MaxBytes = 16 << 20

// MaxArtifactBytes bounds expanded syntax/type evidence, independently of input.
const MaxArtifactBytes = 16 << 20

// Service validates uploads, pins memberships and publishes syntax evidence.
type Service struct {
	// Store supplies the typed, workspace-bound persistence port.
	Store Repository
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID
	// Syntax optionally provides non-Go AST evidence; nil explicitly means unavailable.
	Syntax SyntaxAnalyzer
}

// Digest returns a SHA-256 digest of deterministic Go JSON encoding.
func Digest(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Ingest accepts only sanitized agent snapshots and atomically publishes every
// immutable dependency. Context cancellation aborts before durable publication.
func (s Service) Ingest(ctx context.Context, c IngestCommand) (IngestResult, error) {
	result := IngestResult{}
	// Untrusted payload is checked before ANY repository access or hashing here.
	if err := agent.ValidateSnapshot(c.Snapshot); err != nil {
		return result, fault.ErrInvalid
	}
	if c.WorkspaceID != s.WorkspaceID {
		return result, fault.ErrForbidden
	}
	encoded, _ := json.Marshal(c)
	if len(encoded) > MaxBytes || len(c.Snapshot.Files) > MaxFiles {
		return result, fault.ErrScopeTooLarge
	}
	r, e := s.Store.Repository(ctx, string(c.RepositoryID))
	if e != nil {
		return result, e
	}
	if r.WorkspaceID != s.WorkspaceID {
		return result, fault.ErrForbidden
	}
	defs, e := s.Store.Projects(ctx)
	if e != nil {
		return result, e
	}
	previous, e := s.Store.Latest(ctx)
	if e != nil && !errors.Is(e, fault.ErrNotFound) {
		return result, e
	}
	p := Publication{ExpectedHead: previous.ID, Blobs: []Blob{}, Artifacts: []Artifact{}}
	p.Snapshot = Metadata{ID: Digest([]string{string(s.WorkspaceID), string(r.ID), c.Snapshot.ID}), AgentSnapshotID: c.Snapshot.ID, WorkspaceID: s.WorkspaceID, RepositoryID: r.ID, Repository: c.Snapshot.Repository, Config: c.Snapshot.Config, Diagnostics: c.Snapshot.Diagnostics}
	files := []projects.SourceFile{}
	contents := map[string]string{}
	v := View{WorkspaceID: s.WorkspaceID, SnapshotID: p.Snapshot.ID, Projects: defs, ProjectRevisions: []domain.ProjectRevisionRef{}, Repositories: []domain.RepositorySnapshot{}, Files: []FileRef{}, Profile: "go-ast-syntax/v2;go-codegraph/v2;chunks/v2;syntax-types/v1;syntax-structure/v1;syntax-default", Status: "partial", Projections: projectionStatus()}
	if s.Syntax != nil {
		v.Profile += ";" + s.Syntax.Profile()
		v.Projections["multilanguage_syntax"] = "partial"
		v.Projections["multilanguage_types"] = "partial"
		v.Projections["syntax_structure"] = "partial"
	} else {
		v.Projections["multilanguage_syntax"] = "unsupported"
		v.Projections["multilanguage_types"] = "unsupported"
		v.Projections["syntax_structure"] = "unsupported"
	}
	for _, old := range previous.Repositories {
		if old.RepositoryID != r.ID {
			v.Repositories = append(v.Repositories, old)
		}
	}
	for _, old := range previous.Files {
		if old.RepositoryID != r.ID {
			v.Files = append(v.Files, old)
		}
	}
	if len(v.Files) > 0 {
		oldView := previous
		oldView.Files = v.Files
		loaded, err := s.Store.Files(ctx, oldView)
		if err != nil {
			return result, err
		}
		for _, f := range loaded {
			contents[f.ID] = f.Content
		}
	}
	for _, f := range c.Snapshot.Files {
		if e := ctx.Err(); e != nil {
			return result, e
		}
		files = append(files, projects.SourceFile{Path: f.Path, Content: []byte(f.Content)})
		blobID := Digest([]string{string(s.WorkspaceID), f.Hash})
		id := Digest([]string{p.Snapshot.ID, f.Path})
		p.Blobs = append(p.Blobs, Blob{ID: blobID, WorkspaceID: s.WorkspaceID, SHA256: f.Hash, Content: f.Content})
		contents[id] = f.Content
		v.Files = append(v.Files, FileRef{ID: id, RepositoryID: r.ID, SnapshotID: p.Snapshot.ID, Path: f.Path, BlobID: blobID, Size: f.Size, Language: f.Language, ProjectIDs: []string{}})
	}
	p.Manifest, e = projects.BuildManifest(s.WorkspaceID, r.ID, files)
	if e != nil {
		return result, fault.ErrInvalid
	}
	p.Snapshot.ManifestFingerprint = p.Manifest.Fingerprint
	v.Repositories = append(v.Repositories, domain.RepositorySnapshot{WorkspaceID: s.WorkspaceID, RepositoryID: r.ID, SnapshotID: domain.SnapshotID(p.Snapshot.ID), Commit: c.Snapshot.Repository.Head, ManifestFingerprint: p.Manifest.Fingerprint})
	sort.Slice(v.Repositories, func(i, j int) bool { return v.Repositories[i].RepositoryID < v.Repositories[j].RepositoryID })
	sort.Slice(v.Files, func(i, j int) bool { return v.Files[i].ID < v.Files[j].ID })
	v.Mixed = len(v.Repositories) > 1
	total := 0
	refs := []projects.FileRef{}
	for _, f := range v.Files {
		total += f.Size
		refs = append(refs, projects.FileRef{RepositoryID: f.RepositoryID, Path: f.Path})
	}
	if len(v.Files) > MaxFiles || total > MaxBytes {
		return result, fault.ErrScopeTooLarge
	}
	membership, e := projects.ResolveMembership(defs, refs)
	if e != nil {
		return result, e
	}
	for i, f := range v.Files {
		v.Files[i].ProjectIDs = []string{}
		for _, id := range membership[projects.FileRef{RepositoryID: f.RepositoryID, Path: f.Path}] {
			v.Files[i].ProjectIDs = append(v.Files[i].ProjectIDs, string(id))
		}
	}
	for _, d := range defs {
		fp, err := projects.DefinitionFingerprint(d)
		if err != nil {
			return result, err
		}
		v.ProjectRevisions = append(v.ProjectRevisions, domain.ProjectRevisionRef{WorkspaceID: s.WorkspaceID, ProjectID: d.Project.ID, ProjectRevisionID: domain.RevisionID(fp)})
	}
	v.ID = Digest(v)
	p.View = v
	// A retry returns the original immutable publication and does not rewind head.
	if old, err := s.Store.View(ctx, v.ID); err == nil {
		return IngestResult{SnapshotID: old.SnapshotID, ViewID: old.ID, Status: old.Status}, nil
	} else if !errors.Is(err, fault.ErrNotFound) {
		return result, err
	}
	chunks, e := buildChunks(ctx, v, contents)
	if e != nil {
		return result, e
	}
	graphs, e := buildGraphShards(ctx, v, contents)
	if e != nil {
		return result, e
	}
	artifactBytes := 0
	for _, f := range v.Files {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		report := analysis.Report{}
		if f.Language == "go" {
			report = analysis.Analyze(f.Path, contents[f.ID])
		}
		reportJSON, _ := json.Marshal(report)
		artifact := Artifact{ID: Digest([]string{v.ID, f.ID}), FileID: f.ID, ViewID: v.ID, Report: reportJSON, Types: []typeinfo.TypeDescriptor{}}
		artifact.Chunks = chunks[f.ID]
		artifact.Graph = graphs[f.ID]
		if s.Syntax != nil && supportedSyntax(f.Language) {
			if err := (syntaxArtifactProcessor{analyzer: s.Syntax}).Process(ctx, v, f, contents[f.ID], &artifact); err != nil {
				return result, err
			}
		}
		for _, projectID := range f.ProjectIDs {
			if f.Language != "go" {
				continue
			}
			scope := typeinfo.TypeScope{WorkspaceID: s.WorkspaceID, ProjectID: domain.ProjectID(projectID), BuildProfileID: "syntax-default", VersionID: v.ID}
			descriptors, err := types.FromGoReport(scope, f.RepositoryID, contents[f.ID], report)
			if err != nil {
				return result, err
			}
			artifact.Types = append(artifact.Types, descriptors...)
		}
		artifactJSON, _ := json.Marshal(artifact)
		artifactBytes += len(artifactJSON)
		if artifactBytes > MaxArtifactBytes {
			return result, fault.ErrScopeTooLarge
		}
		p.Artifacts = append(p.Artifacts, artifact)
	}
	if e = s.Store.Publish(ctx, p); e != nil {
		return result, e
	}
	return IngestResult{SnapshotID: p.Snapshot.ID, ViewID: v.ID, Status: v.Status}, nil
}
