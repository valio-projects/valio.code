// Package domain defines durable identities and immutable analysis references.
package domain

import (
	"fmt"
	"path"
	"strings"
	"time"
)

// WorkspaceID identifies an isolated workspace.
type WorkspaceID string

// ProjectID identifies a project within a workspace.
type ProjectID string

// RepositoryID identifies a configured source repository.
type RepositoryID string

// WorktreeID identifies a local checkout of a repository.
type WorktreeID string

// RevisionID identifies an append-only project definition revision.
type RevisionID string

// SnapshotID identifies immutable repository or source inputs.
type SnapshotID string

// ViewID identifies an immutable analysis view.
type ViewID string

// Workspace is an identity and display name for an isolated collection of projects.
type Workspace struct {
	// ID is the stable workspace identity.
	ID WorkspaceID `json:"id"`
	// Name is the user-facing workspace name.
	Name string `json:"name"`
}

// ProjectKind classifies the runtime purpose of a project.
type ProjectKind string

const (
	ProjectService     ProjectKind = "service"
	ProjectLibrary     ProjectKind = "library"
	ProjectApplication ProjectKind = "application"
	ProjectTool        ProjectKind = "tool"
)

// Valid reports whether k is a supported project classification.
func (k ProjectKind) Valid() bool {
	return k == ProjectService || k == ProjectLibrary || k == ProjectApplication || k == ProjectTool
}

// Project contains stable identity and editable definition metadata.
type Project struct {
	// ID is the stable project identity.
	ID ProjectID `json:"id"`
	// WorkspaceID scopes the project identity and key.
	WorkspaceID WorkspaceID `json:"workspaceId"`
	// Key is an immutable workspace-unique slug.
	Key string `json:"key"`
	// Name is the editable display name.
	Name string `json:"name"`
	// Description is optional user-facing context.
	Description string `json:"description"`
	// Kind classifies the project.
	Kind ProjectKind `json:"kind"`
	// Tags are unique, trimmed labels.
	Tags []string `json:"tags"`
	// BuildProfiles names applicable build configurations.
	BuildProfiles []string `json:"buildProfiles"`
	// EnvironmentProfiles names applicable deployment configurations.
	EnvironmentProfiles []string `json:"environmentProfiles"`
	// Status controls whether the project is active or archived.
	Status ProjectStatus `json:"status"`
}

// ProjectStatus records whether a project participates in normal work.
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

// ValidateProjectKey validates a stable, workspace-unique slug. Persistence must
// reject changing the key of an existing identity, including during renames.
func ValidateProjectKey(key string) error {
	if len(key) < 1 || len(key) > 80 {
		return fmt.Errorf("project key must contain 1 to 80 characters")
	}
	for i, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c == '-' && i > 0 && i < len(key)-1)) {
			return fmt.Errorf("project key must be a lowercase alphanumeric slug with internal hyphens")
		}
	}
	return nil
}

// Validate checks required project identity, classification, key, status, and set-like labels.
func (p Project) Validate() error {
	if p.ID == "" || p.WorkspaceID == "" || strings.TrimSpace(p.Name) == "" || !p.Kind.Valid() {
		return fmt.Errorf("project requires identity, workspace, name and valid kind")
	}
	if err := ValidateProjectKey(p.Key); err != nil {
		return err
	}
	if p.Status != ProjectActive && p.Status != ProjectArchived {
		return fmt.Errorf("invalid project status %q", p.Status)
	}
	for _, values := range [][]string{p.Tags, p.BuildProfiles, p.EnvironmentProfiles} {
		seen := map[string]bool{}
		for _, value := range values {
			if strings.TrimSpace(value) == "" || strings.TrimSpace(value) != value || seen[value] {
				return fmt.Errorf("tags and profile identifiers must be nonempty, trimmed and unique")
			}
			seen[value] = true
		}
	}
	return nil
}

// ValidateUpdate permits display-name changes while preserving stable identity.
func (p Project) ValidateUpdate(previous Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.ID != previous.ID || p.WorkspaceID != previous.WorkspaceID || p.Key != previous.Key {
		return fmt.Errorf("project identity, workspace and key are immutable")
	}
	return nil
}

// Repository identifies a workspace-scoped source remote.
type Repository struct {
	// ID is the stable repository identity.
	ID RepositoryID `json:"id"`
	// WorkspaceID scopes access to the repository.
	WorkspaceID WorkspaceID `json:"workspaceId"`
	// RemoteURL is the configured remote location.
	RemoteURL string `json:"remoteUrl"`
}

// Worktree records a local checkout state; it is distinct from a repository.
type Worktree struct {
	// ID is the worktree identity.
	ID WorktreeID `json:"id"`
	// RepositoryID identifies the checkout's source repository.
	RepositoryID RepositoryID `json:"repositoryId"`
	// LocalPath is the machine-local checkout location.
	LocalPath string `json:"localPath"`
	// HeadCommit is the checked-out commit when known.
	HeadCommit string `json:"headCommit"`
	// Dirty reports uncommitted working-tree changes.
	Dirty bool `json:"dirty"`
}

// ProjectSourceRoot relates a project to a repository. Roots may overlap across
// and within projects. Include/exclude globs are relative to Path.
type ProjectSourceRoot struct {
	RepositoryID RepositoryID   `json:"repositoryId"`
	Path         string         `json:"path"`
	Include      []string       `json:"include,omitempty"`
	Exclude      []string       `json:"exclude,omitempty"`
	Role         SourceRootRole `json:"role"`
	BuildUnit    string         `json:"buildUnit,omitempty"`
	Version      string         `json:"version"`
}

type SourceRootRole string

const (
	RootCode          SourceRootRole = "code"
	RootTests         SourceRootRole = "tests"
	RootContracts     SourceRootRole = "contracts"
	RootConfiguration SourceRootRole = "configuration"
	RootDeployment    SourceRootRole = "deployment"
	RootDocumentation SourceRootRole = "documentation"
	RootGenerated     SourceRootRole = "generated"
)

// EffectiveRole keeps an omitted role equivalent to code.
func (r ProjectSourceRoot) EffectiveRole() SourceRootRole {
	if r.Role == "" {
		return RootCode
	}
	return r.Role
}
func (r ProjectSourceRoot) Validate() error {
	if r.RepositoryID == "" || strings.TrimSpace(r.Version) == "" {
		return fmt.Errorf("source root requires repository identity and version")
	}
	switch r.EffectiveRole() {
	case RootCode, RootTests, RootContracts, RootConfiguration, RootDeployment, RootDocumentation, RootGenerated:
	default:
		return fmt.Errorf("invalid source root role %q", r.Role)
	}
	if strings.TrimSpace(r.BuildUnit) != r.BuildUnit {
		return fmt.Errorf("build unit identifier must be trimmed")
	}
	return ValidateRelativePath(r.Path, true)
}

// ProjectRevision is append-only: any definition change creates another ID.
// Persistence implementations must reject replacement of an existing revision.
type ProjectRevision struct {
	ID                    RevisionID `json:"id"`
	ProjectID             ProjectID  `json:"projectId"`
	DefinitionFingerprint string     `json:"definitionFingerprint"`
	CreatedAt             time.Time  `json:"createdAt"`
}

func (r ProjectRevision) Validate() error {
	if r.ID == "" || r.ProjectID == "" || r.DefinitionFingerprint == "" || r.CreatedAt.IsZero() {
		return fmt.Errorf("project revision requires identity, project, definition fingerprint and creation time")
	}
	return nil
}

// RepositorySnapshot pins an input to its content manifest, including dirty
// worktree files. A mutable branch name alone is never an analysis reference.
type RepositorySnapshot struct {
	WorkspaceID         WorkspaceID  `json:"workspaceId"`
	SnapshotID          SnapshotID   `json:"snapshotId"`
	RepositoryID        RepositoryID `json:"repositoryId"`
	WorktreeID          WorktreeID   `json:"worktreeId,omitempty"`
	Commit              string       `json:"commit,omitempty"`
	ManifestFingerprint string       `json:"manifestFingerprint"`
}
type SourceSnapshot struct {
	ID           SnapshotID           `json:"id"`
	WorkspaceID  WorkspaceID          `json:"workspaceId"`
	Repositories []RepositorySnapshot `json:"repositories"`
	CreatedAt    time.Time            `json:"createdAt"`
}

func (s SourceSnapshot) Validate() error {
	if s.ID == "" || s.WorkspaceID == "" || s.CreatedAt.IsZero() || len(s.Repositories) == 0 {
		return fmt.Errorf("source snapshot requires identity, workspace, creation time and repository inputs")
	}
	seen := map[RepositoryID]bool{}
	for _, r := range s.Repositories {
		if r.WorkspaceID != s.WorkspaceID {
			return fmt.Errorf("repository snapshot workspace differs from source snapshot")
		}
		if r.SnapshotID == "" || r.RepositoryID == "" || r.ManifestFingerprint == "" {
			return fmt.Errorf("repository snapshot requires repository identity and concrete manifest fingerprint")
		}
		if seen[r.RepositoryID] {
			return fmt.Errorf("duplicate repository snapshot %q", r.RepositoryID)
		}
		seen[r.RepositoryID] = true
	}
	return nil
}

type ProjectRevisionRef struct {
	WorkspaceID       WorkspaceID `json:"workspaceId"`
	ProjectID         ProjectID   `json:"projectId"`
	ProjectRevisionID RevisionID  `json:"projectRevisionId"`
}
type ProjectionGenerationRef struct {
	WorkspaceID  WorkspaceID `json:"workspaceId"`
	Family       string      `json:"family"`
	GenerationID string      `json:"generationId"`
}

// AnalysisView is a workspace-scoped immutable set of concrete project revisions,
// repository inputs and projection generations. A single view can span projects.
type AnalysisView struct {
	ID                    ViewID                    `json:"id"`
	WorkspaceID           WorkspaceID               `json:"workspaceId"`
	Projects              []ProjectRevisionRef      `json:"projects"`
	SourceSnapshotID      SnapshotID                `json:"sourceSnapshotId"`
	Repositories          []RepositorySnapshot      `json:"repositories"`
	ProjectionGenerations []ProjectionGenerationRef `json:"projectionGenerations"`
	ProfileFingerprint    string                    `json:"profileFingerprint"`
}

func (v AnalysisView) Validate() error {
	if v.ID == "" || v.WorkspaceID == "" || len(v.Projects) == 0 || len(v.Repositories) == 0 || len(v.ProjectionGenerations) == 0 || v.SourceSnapshotID == "" || v.ProfileFingerprint == "" {
		return fmt.Errorf("analysis view requires workspace and concrete project revision, repository snapshot, projection generation and profile references")
	}
	projects := map[ProjectID]bool{}
	for _, p := range v.Projects {
		if p.WorkspaceID != v.WorkspaceID || p.ProjectID == "" || p.ProjectRevisionID == "" {
			return fmt.Errorf("invalid project revision reference or mixed workspace")
		}
		if projects[p.ProjectID] {
			return fmt.Errorf("duplicate project reference %q", p.ProjectID)
		}
		projects[p.ProjectID] = true
	}
	repositories := map[RepositoryID]bool{}
	for _, r := range v.Repositories {
		if r.WorkspaceID != v.WorkspaceID || r.SnapshotID == "" || r.RepositoryID == "" || r.ManifestFingerprint == "" {
			return fmt.Errorf("invalid repository snapshot reference or mixed workspace")
		}
		if repositories[r.RepositoryID] {
			return fmt.Errorf("duplicate repository reference %q", r.RepositoryID)
		}
		repositories[r.RepositoryID] = true
	}
	families := map[string]bool{}
	for _, p := range v.ProjectionGenerations {
		if p.WorkspaceID != v.WorkspaceID || p.Family == "" || p.GenerationID == "" {
			return fmt.Errorf("invalid projection generation reference or mixed workspace")
		}
		if families[p.Family] {
			return fmt.Errorf("duplicate projection family %q", p.Family)
		}
		families[p.Family] = true
	}
	return nil
}

// ValidateRelativePath accepts slash-separated, clean repository-relative
// paths. The root is represented only by "." when allowRoot is true. It rejects
// Windows paths even when running on Unix.
func ValidateRelativePath(p string, allowRoot bool) error {
	if allowRoot && p == "." {
		return nil
	}
	if p == "" || p == "." || strings.ContainsAny(p, "\\:\x00") || strings.HasPrefix(p, "/") || path.Clean(p) != p {
		return fmt.Errorf("invalid portable relative path %q", p)
	}
	for _, segment := range strings.Split(p, "/") {
		if segment == ".." || segment == "." || strings.TrimSpace(segment) == "" || strings.ContainsAny(segment, "<>\"|?*") || strings.HasSuffix(segment, ".") || strings.HasSuffix(segment, " ") {
			return fmt.Errorf("invalid portable path segment in %q", p)
		}
		base := strings.ToUpper(strings.SplitN(segment, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return fmt.Errorf("reserved Windows path segment in %q", p)
		}
		for _, c := range segment {
			if c < 32 {
				return fmt.Errorf("control character in path %q", p)
			}
		}
	}
	return nil
}
