// Package git provides read-only Git discovery without invoking a shell.
package git

import (
	"context"
	"errors"
	"github.com/valio-projects/valio.code/internal/validation"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Repository struct {
	Root      string     `json:"root"`
	GitDir    string     `json:"gitDir"`
	CommonDir string     `json:"commonDir"`
	Head      string     `json:"head,omitempty"`
	Branch    string     `json:"branch,omitempty"`
	Refs      []string   `json:"refs"`
	Dirty     bool       `json:"dirty"`
	Worktrees []Worktree `json:"worktrees"`
	Status    []Status   `json:"status"`
}

type Worktree struct {
	Path     string `json:"path"`
	Head     string `json:"head,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Bare     bool   `json:"bare,omitempty"`
	Detached bool   `json:"detached,omitempty"`
}
type Status struct {
	Path         string `json:"path"`
	Index        string `json:"index"`
	Worktree     string `json:"worktree"`
	OriginalPath string `json:"originalPath,omitempty"`
}

// Run never includes Git stderr in errors: filenames and configured helpers can
// contain sensitive user data. Git hooks, pagers and optional lock writes are off.
func Run(ctx context.Context, root string, args ...string) ([]byte, error) {
	argv := append([]string{"--no-pager", "--no-optional-locks", "-c", "core.fsmonitor=false", "-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", argv...)
	// Ambient Git overrides must not redirect discovery to a different index
	// or worktree. Configured global ignore rules are still respected by Git.
	for _, env := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(env), "GIT_") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("git operation failed")
	}
	return out, nil
}

func Discover(ctx context.Context, root string) (Repository, error) {
	abs, err := validation.Directory(root)
	if err != nil {
		return Repository{}, err
	}
	b, err := Run(ctx, abs, "rev-parse", "--show-toplevel")
	if err != nil {
		if ctx.Err() != nil {
			return Repository{}, ctx.Err()
		}
		return Repository{}, errors.New("root is not a Git worktree")
	}
	r := Repository{Root: filepath.Clean(strings.TrimSpace(string(b))), Refs: []string{}}
	for _, item := range []struct {
		arg    string
		target *string
	}{{"--absolute-git-dir", &r.GitDir}, {"--git-common-dir", &r.CommonDir}, {"HEAD", &r.Head}} {
		b, e := Run(ctx, r.Root, "rev-parse", item.arg)
		if e == nil {
			*item.target = strings.TrimSpace(string(b))
		}
	}
	if r.CommonDir != "" && !filepath.IsAbs(r.CommonDir) {
		r.CommonDir = filepath.Join(r.Root, r.CommonDir)
	}
	b, _ = Run(ctx, r.Root, "symbolic-ref", "--quiet", "--short", "HEAD")
	r.Branch = strings.TrimSpace(string(b))
	b, err = Run(ctx, r.Root, "for-each-ref", "--format=%(refname)")
	if err != nil {
		return Repository{}, err
	}
	for _, s := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if s != "" {
			r.Refs = append(r.Refs, s)
		}
	}
	sort.Strings(r.Refs)
	b, err = Run(ctx, r.Root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return Repository{}, err
	}
	r.Dirty = len(b) > 0
	r.Status = parseStatus(b)
	b, err = Run(ctx, r.Root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return Repository{}, err
	}
	r.Worktrees = parseWorktrees(b)
	return r, nil
}

func parseStatus(data []byte) []Status {
	result := []Status{}
	parts := strings.Split(string(data), "\x00")
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if len(p) < 4 {
			continue
		}
		s := Status{Path: p[3:], Index: p[:1], Worktree: p[1:2]}
		if strings.ContainsAny(p[:2], "RC") && i+1 < len(parts) {
			i++
			s.OriginalPath = parts[i]
		}
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
func parseWorktrees(data []byte) []Worktree {
	result := []Worktree{}
	var current *Worktree
	for _, p := range strings.Split(string(data), "\x00") {
		if strings.HasPrefix(p, "worktree ") {
			result = append(result, Worktree{Path: strings.TrimPrefix(p, "worktree ")})
			current = &result[len(result)-1]
			continue
		}
		if current == nil {
			continue
		}
		switch {
		case strings.HasPrefix(p, "HEAD "):
			current.Head = strings.TrimPrefix(p, "HEAD ")
		case strings.HasPrefix(p, "branch "):
			current.Branch = strings.TrimPrefix(p, "branch ")
		case p == "bare":
			current.Bare = true
		case p == "detached":
			current.Detached = true
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

// Files combines tracked and untracked paths using Git's own ignore rules.
// Tracked files remain present even if later covered by .gitignore.
func Files(ctx context.Context, root string) ([]string, error) {
	b, err := Run(ctx, root, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	paths := []string{}
	for _, s := range strings.Split(string(b), "\x00") {
		if s != "" && !seen[s] {
			seen[s] = true
			paths = append(paths, s)
		}
	}
	sort.Strings(paths)
	return paths, nil
}
