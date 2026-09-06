package projects

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/valio-projects/valio.code/internal/domain"
)

// BlobRef scopes content-addressed bytes to a workspace; digest equality never
// grants another workspace access to the blob.
type BlobRef struct {
	// WorkspaceID scopes the digest to an access boundary.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// SHA256 is the lowercase hexadecimal content digest.
	SHA256 string `json:"sha256"`
}

// SourceFile supplies repository-relative path and raw bytes for manifest construction.
type SourceFile struct {
	// Path is a portable repository-relative path.
	Path string
	// Content is the exact byte content to hash.
	Content []byte
}

// ManifestEntry binds a path to a scoped blob and its byte size.
type ManifestEntry struct {
	// Path is the repository-relative file path.
	Path string `json:"path"`
	// Blob identifies immutable file content.
	Blob BlobRef `json:"blob"`
	// Size is the file length in bytes.
	Size int `json:"size"`
}

// Manifest is a deterministic repository input inventory split into hash shards.
type Manifest struct {
	// WorkspaceID scopes all entries and the fingerprint.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// RepositoryID identifies the repository represented.
	RepositoryID domain.RepositoryID `json:"repositoryId"`
	// Entries contains files sorted by path.
	Entries []ManifestEntry `json:"entries"`
	// Shards contains fingerprints indexed by the first path-hash byte.
	Shards [256]string `json:"shards"`
	// Fingerprint identifies this complete workspace-scoped inventory.
	Fingerprint string `json:"fingerprint"`
}

// BuildManifest sorts files and hashes each path/content pair into one of 256
// shards selected by the first SHA-256 byte of its path. File order is irrelevant.
func BuildManifest(workspace domain.WorkspaceID, repository domain.RepositoryID, files []SourceFile) (Manifest, error) {
	builder, err := NewManifestBuilder(workspace, repository)
	if err != nil {
		return Manifest{}, err
	}
	return builder.Build(files)
}

// Build validates files and returns a deterministic manifest without retaining file bytes.
func (b *ManifestBuilder) Build(files []SourceFile) (Manifest, error) {
	if b == nil {
		return Manifest{}, fmt.Errorf("manifest builder is required")
	}
	workspace, repository := b.workspace, b.repository
	if workspace == "" || repository == "" {
		return Manifest{}, fmt.Errorf("manifest requires workspace and repository identities")
	}
	m := Manifest{WorkspaceID: workspace, RepositoryID: repository, Entries: []ManifestEntry{}}
	seen := map[string]bool{}
	for _, f := range files {
		if err := domain.ValidateRelativePath(f.Path, false); err != nil {
			return Manifest{}, err
		}
		if seen[f.Path] {
			return Manifest{}, fmt.Errorf("duplicate file %q", f.Path)
		}
		seen[f.Path] = true
		h := sha256.Sum256(f.Content)
		m.Entries = append(m.Entries, ManifestEntry{Path: f.Path, Blob: BlobRef{workspace, hex.EncodeToString(h[:])}, Size: len(f.Content)})
	}
	sort.Slice(m.Entries, func(i, j int) bool { return m.Entries[i].Path < m.Entries[j].Path })
	var shards [256][]ManifestEntry
	for _, e := range m.Entries {
		h := sha256.Sum256([]byte(e.Path))
		shards[h[0]] = append(shards[h[0]], e)
	}
	for i := range shards {
		m.Shards[i] = digest(shards[i])
	}
	m.Fingerprint = digest(struct {
		Version    string
		Workspace  domain.WorkspaceID
		Repository domain.RepositoryID
		Shards     [256]string
	}{"source-manifest/v1", workspace, repository, m.Shards})
	return m, nil
}

// Fingerprint uses only files included in this project. An unrelated path in a
// monorepo, or an unrelated repository, cannot invalidate the project input.
func Fingerprint(d Definition, manifests []Manifest) (string, error) {
	definition, err := DefinitionFingerprint(d)
	if err != nil {
		return "", err
	}
	type input struct {
		Repository domain.RepositoryID
		Entry      ManifestEntry
	}
	selected := []input{}
	seen := map[domain.RepositoryID]bool{}
	for _, m := range manifests {
		if seen[m.RepositoryID] {
			return "", fmt.Errorf("duplicate repository manifest %q", m.RepositoryID)
		}
		seen[m.RepositoryID] = true
		if m.WorkspaceID != d.Project.WorkspaceID {
			return "", fmt.Errorf("manifest workspace differs from project")
		}
		for _, e := range m.Entries {
			if contains(d, FileRef{m.RepositoryID, e.Path}) {
				selected = append(selected, input{m.RepositoryID, e})
			}
		}
	}
	for _, root := range d.Roots {
		if !seen[root.RepositoryID] {
			return "", fmt.Errorf("missing repository manifest %q", root.RepositoryID)
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Repository != selected[j].Repository {
			return selected[i].Repository < selected[j].Repository
		}
		return selected[i].Entry.Path < selected[j].Entry.Path
	})
	return digest(struct {
		Version, Definition string
		Files               []input
	}{"project-source/v1", definition, selected}), nil
}
