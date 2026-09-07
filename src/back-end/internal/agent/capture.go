// Package agent captures immutable, sanitized local repository snapshots.
package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/configgraph"
	gitrepo "github.com/valio-projects/valio.code/internal/git"
)

// DefaultMaxFileBytes is the largest source file retained as content: 2 MiB.
const DefaultMaxFileBytes int64 = 2 * 1024 * 1024

// Options configures one local repository capture.
type Options struct {
	// Root is the repository worktree to discover.
	Root string
	// SpoolDir optionally persists the validated snapshot.
	SpoolDir string
	// MaxFileBytes bounds a source file; values above the default are clamped.
	MaxFileBytes int64
}

// File is a sanitized source file retained in a snapshot.
type File struct {
	// Path is a repository-relative path.
	Path string `json:"path"`
	// Hash is the SHA-256 digest of Content.
	Hash string `json:"hash"`
	// Size is the content length in bytes.
	Size int `json:"size"`
	// Language is an extension-derived classification.
	Language string `json:"language"`
	// Content passed all capture policy checks.
	Content string `json:"content"`
}

// Diagnostic records a safe reason for an excluded path.
type Diagnostic struct {
	// Path is the affected repository-relative path when safe to report.
	Path string `json:"path,omitempty"`
	// Code is a stable capture-policy outcome.
	Code string `json:"code"`
	// Message is a fixed explanation without source contents.
	Message string `json:"message"`
}

// Snapshot is a content-addressed, sanitized repository capture.
type Snapshot struct {
	// Version identifies the serialization contract.
	Version int `json:"version"`
	// ID hashes this snapshot with ID omitted.
	ID string `json:"id"`
	// Repository contains bounded Git metadata or an explicit unknown state.
	Repository gitrepo.Repository `json:"repository"`
	// Files contains accepted non-configuration files.
	Files []File `json:"files"`
	// Config contains key-only configuration projections.
	Config []configgraph.Projection `json:"config"`
	// Diagnostics explains exclusions without exposing content.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CaptureBuilder combines options with optional reading and persistence adapters.
type CaptureBuilder struct {
	// Options controls root, spool, and source-size budget.
	Options Options
	// Reader optionally replaces filesystem reads after policy filtering.
	Reader SourceReader
	// Writer optionally receives the validated snapshot.
	Writer SnapshotWriter
}

// Capture discovers opts.Root and returns a validated sanitized snapshot.
func Capture(ctx context.Context, opts Options) (Snapshot, error) {
	return (&CaptureBuilder{Options: opts}).Build(ctx)
}

// Build captures through configured adapters and never reads excluded paths.
func (builder *CaptureBuilder) Build(ctx context.Context) (Snapshot, error) {
	opts := builder.Options
	repo, err := gitrepo.Discover(ctx, opts.Root)
	if err != nil {
		return Snapshot{}, err
	}
	root, err := filepath.EvalSymlinks(repo.Root)
	if err != nil {
		return Snapshot{}, errors.New("cannot resolve repository root")
	}
	repo.Root = root
	rootFS, err := os.OpenRoot(root)
	if err != nil {
		return Snapshot{}, errors.New("cannot open repository root")
	}
	defer rootFS.Close()
	reader := builder.Reader
	if reader == nil {
		reader = FilesystemReader{Root: rootFS}
	}
	paths, err := gitrepo.Files(ctx, root)
	if err != nil {
		return Snapshot{}, err
	}
	limit := opts.MaxFileBytes
	if limit <= 0 || limit > DefaultMaxFileBytes {
		limit = DefaultMaxFileBytes
	}
	rules, err := readIgnore(root)
	if err != nil {
		return Snapshot{}, err
	}
	// Agent-owned spool files must not change the identity they are storing.
	filteredStatus := repo.Status[:0]
	for _, status := range repo.Status {
		if status.Path == ".valio" || strings.HasPrefix(status.Path, ".valio/") {
			continue
		}
		if opts.SpoolDir != "" {
			spool, _ := filepath.Abs(opts.SpoolDir)
			if within(spool, filepath.Join(root, filepath.FromSlash(status.Path))) {
				continue
			}
		}
		filteredStatus = append(filteredStatus, status)
	}
	repo.Status = filteredStatus
	repo.Dirty = len(filteredStatus) > 0
	s := Snapshot{Version: 1, Repository: repo, Files: []File{}, Config: []configgraph.Projection{}, Diagnostics: []Diagnostic{}}
	diag := func(path, code, message string) {
		s.Diagnostics = append(s.Diagnostics, Diagnostic{path, code, message})
	}
	for _, path := range paths {
		if !safeRelativePath(path) {
			diag("", "unsafe-path", "path leaves repository root")
			continue
		}
		if err := ctx.Err(); err != nil {
			return Snapshot{}, errors.New("capture canceled")
		}
		if path == ".git" || strings.HasPrefix(path, ".git/") || path == ".valio" || strings.HasPrefix(path, ".valio/") {
			continue
		}
		if ignored(path, rules) {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(path))
		if !within(root, full) {
			diag(path, "unsafe-path", "path leaves repository root")
			continue
		}
		if opts.SpoolDir != "" {
			spool, _ := filepath.Abs(opts.SpoolDir)
			if within(spool, full) {
				continue
			}
		}
		info, e := os.Lstat(full)
		if e != nil {
			diag(path, "unreadable", "file missing or unreadable")
			continue
		}
		// Never follow symlinks, including links in parent directories.
		resolved, e := filepath.EvalSymlinks(full)
		if e != nil || !within(root, resolved) || !samePath(full, resolved) || info.Mode()&os.ModeSymlink != 0 {
			diag(path, "symlink-excluded", "symlink paths are excluded")
			continue
		}
		if !info.Mode().IsRegular() {
			diag(path, "nonregular-excluded", "non-regular file excluded")
			continue
		}
		if sensitiveName(path) {
			diag(path, "sensitive-excluded", "sensitive artifact excluded before reading")
			continue
		}
		if info.Size() > limit {
			diag(path, "oversize-excluded", "file exceeds capture size limit")
			continue
		}
		// os.Root prevents a concurrent symlink replacement from escaping the
		// repository between validation and open.
		data, e := reader.Read(filepath.FromSlash(path), limit)
		if e != nil {
			diag(path, "unreadable", "file cannot be read within size limit")
			continue
		}
		if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
			diag(path, "binary-excluded", "binary or non-UTF-8 content excluded")
			continue
		}
		kind := configgraph.Kind(path)
		if kind != "" {
			s.Config = append(s.Config, configgraph.Extract(path, kind, data))
			diag(path, "config-projected", "configuration values excluded; only safe key metadata retained")
			continue
		}
		if suspiciousContent(data) {
			diag(path, "secret-pattern-excluded", "content containing a potential credential excluded before hashing")
			continue
		}
		hash := sha256.Sum256(data)
		s.Files = append(s.Files, File{Path: path, Hash: hex.EncodeToString(hash[:]), Size: len(data), Language: language(path), Content: string(data)})
	}
	sort.Slice(s.Diagnostics, func(i, j int) bool {
		if s.Diagnostics[i].Path == s.Diagnostics[j].Path {
			return s.Diagnostics[i].Code < s.Diagnostics[j].Code
		}
		return s.Diagnostics[i].Path < s.Diagnostics[j].Path
	})
	encoded, err := json.Marshal(s)
	if err != nil {
		return Snapshot{}, errors.New("cannot encode snapshot")
	}
	sum := sha256.Sum256(encoded)
	s.ID = hex.EncodeToString(sum[:])
	if err := ValidateSnapshot(s); err != nil {
		return Snapshot{}, err
	}
	writer := builder.Writer
	if writer == nil && opts.SpoolDir != "" {
		writer = DiskSpool{Dir: opts.SpoolDir}
	}
	if writer != nil {
		if err := writer.Write(s); err != nil {
			return Snapshot{}, err
		}
	}
	return s, nil
}

func within(root, path string) bool {
	rel, e := filepath.Rel(root, path)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func samePath(a, b string) bool { return strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) }
