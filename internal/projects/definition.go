// Package projects validates project definitions and resolves their many-to-many
// membership independently of repository/worktree identities.
package projects

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/valio-projects/valio.code/internal/domain"
)

type Definition struct {
	Project domain.Project             `json:"project"`
	Roots   []domain.ProjectSourceRoot `json:"roots"`
}

// Globs are anchored to the source root. Each component supports path.Match
// syntax (*, ?, [class]); ** as an entire component matches zero or more path
// components. Empty Include selects all files; Exclude always takes precedence.
func ValidateGlob(pattern string) error {
	if pattern == "" || strings.HasPrefix(pattern, "/") || strings.ContainsAny(pattern, "\\:\x00") {
		return fmt.Errorf("invalid rooted glob %q", pattern)
	}
	for _, segment := range strings.Split(pattern, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("unclean glob %q", pattern)
		}
		if strings.Contains(segment, "**") && segment != "**" {
			return fmt.Errorf("** must occupy an entire glob component: %q", pattern)
		}
		if segment != "**" {
			if _, err := path.Match(segment, ""); err != nil {
				return fmt.Errorf("invalid glob %q: %w", pattern, err)
			}
		}
	}
	return nil
}

func (d Definition) Validate() error {
	if err := d.Project.Validate(); err != nil {
		return err
	}
	if len(d.Roots) == 0 {
		return fmt.Errorf("project requires at least one source root")
	}
	for _, root := range d.Roots {
		if err := root.Validate(); err != nil {
			return err
		}
		for _, patterns := range [][]string{root.Include, root.Exclude} {
			for _, p := range patterns {
				if err := ValidateGlob(p); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func matchGlob(pattern, relative string) bool {
	p, n := strings.Split(pattern, "/"), strings.Split(relative, "/")
	// Dynamic programming avoids exponential backtracking with repeated **.
	type key struct{ i, j int }
	memo, seen := map[key]bool{}, map[key]bool{}
	var match func(int, int) bool
	match = func(i, j int) bool {
		k := key{i, j}
		if seen[k] {
			return memo[k]
		}
		seen[k] = true
		ok := false
		if i == len(p) {
			ok = j == len(n)
		} else if p[i] == "**" {
			ok = match(i+1, j) || (j < len(n) && match(i, j+1))
		} else if j < len(n) {
			m, _ := path.Match(p[i], n[j])
			ok = m && match(i+1, j+1)
		}
		memo[k] = ok
		return ok
	}
	return match(0, 0)
}

type FileRef struct {
	RepositoryID domain.RepositoryID `json:"repositoryId"`
	Path         string              `json:"path"`
}

func contains(d Definition, file FileRef) bool {
	for _, root := range d.Roots {
		if file.RepositoryID != root.RepositoryID {
			continue
		}
		rel := file.Path
		if root.Path != "." {
			if !strings.HasPrefix(file.Path, root.Path+"/") {
				continue
			}
			rel = strings.TrimPrefix(file.Path, root.Path+"/")
		}
		included := len(root.Include) == 0
		for _, p := range root.Include {
			if matchGlob(p, rel) {
				included = true
				break
			}
		}
		if !included {
			continue
		}
		for _, p := range root.Exclude {
			if matchGlob(p, rel) {
				included = false
				break
			}
		}
		if included {
			return true
		}
	}
	return false
}

// ResolveMembership returns each input file's sorted project IDs, including an
// empty slice for files outside all projects. No ownership winner is imposed.
func ResolveMembership(definitions []Definition, files []FileRef) (map[FileRef][]domain.ProjectID, error) {
	resolver, err := NewMembershipResolver(definitions)
	if err != nil {
		return nil, err
	}
	return resolver.Resolve(files)
}

func resolveMembership(definitions []Definition, files []FileRef) (map[FileRef][]domain.ProjectID, error) {
	ids := map[domain.ProjectID]bool{}
	type workspaceKey struct {
		workspace domain.WorkspaceID
		key       string
	}
	keys := map[workspaceKey]bool{}
	for _, d := range definitions {
		if err := d.Validate(); err != nil {
			return nil, err
		}
		if ids[d.Project.ID] {
			return nil, fmt.Errorf("duplicate project %q", d.Project.ID)
		}
		ids[d.Project.ID] = true
		key := workspaceKey{d.Project.WorkspaceID, d.Project.Key}
		if keys[key] {
			return nil, fmt.Errorf("duplicate project key %q in workspace %q", d.Project.Key, d.Project.WorkspaceID)
		}
		keys[key] = true
	}
	out := make(map[FileRef][]domain.ProjectID, len(files))
	for _, f := range files {
		if f.RepositoryID == "" {
			return nil, fmt.Errorf("file requires repository identity")
		}
		if err := domain.ValidateRelativePath(f.Path, false); err != nil {
			return nil, err
		}
		members := []domain.ProjectID{}
		for _, d := range definitions {
			if contains(d, f) {
				members = append(members, d.Project.ID)
			}
		}
		sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
		out[f] = members
	}
	return out, nil
}

func digest(value any) string {
	b, _ := json.Marshal(value)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func sortedUnique(values []string) []string {
	v := append([]string{}, values...)
	sort.Strings(v)
	out := make([]string, 0, len(v))
	for _, s := range v {
		if len(out) == 0 || out[len(out)-1] != s {
			out = append(out, s)
		}
	}
	return out
}
func DefinitionFingerprint(d Definition) (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}
	d.Project.Tags = sortedUnique(d.Project.Tags)
	d.Project.BuildProfiles = sortedUnique(d.Project.BuildProfiles)
	d.Project.EnvironmentProfiles = sortedUnique(d.Project.EnvironmentProfiles)
	// Normalize set-like roots and patterns, without mutating caller data.
	rootKeys := []string{}
	for _, r := range d.Roots {
		r.Role = r.EffectiveRole()
		r.Include = sortedUnique(r.Include)
		r.Exclude = sortedUnique(r.Exclude)
		b, _ := json.Marshal(r)
		rootKeys = append(rootKeys, string(b))
	}
	return digest(struct {
		Version string
		Project domain.Project
		Roots   []string
	}{"project-definition/v1", d.Project, sortedUnique(rootKeys)}), nil
}

// ProfileFingerprint includes exact analyzer versions/configuration. The map
// encoding is canonical because encoding/json sorts string map keys.
func ProfileFingerprint(versions, configuration map[string]string) string {
	return digest(struct {
		Version                  string
		Analyzers, Configuration map[string]string
	}{"analysis-profile/v1", versions, configuration})
}
