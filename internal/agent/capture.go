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

const DefaultMaxFileBytes int64 = 2 * 1024 * 1024

type Options struct {
	Root         string
	SpoolDir     string
	MaxFileBytes int64
}
type File struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	Size     int    `json:"size"`
	Language string `json:"language"`
	Content  string `json:"content"`
}
type Diagnostic struct {
	Path    string `json:"path,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Snapshot struct {
	Version     int                      `json:"version"`
	ID          string                   `json:"id"`
	Repository  gitrepo.Repository       `json:"repository"`
	Files       []File                   `json:"files"`
	Config      []configgraph.Projection `json:"config"`
	Diagnostics []Diagnostic             `json:"diagnostics"`
}

type CaptureBuilder struct {
	Options Options
	Reader  SourceReader
	Writer  SnapshotWriter
}

func Capture(ctx context.Context, opts Options) (Snapshot, error) {
	return (&CaptureBuilder{Options: opts}).Build(ctx)
}

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
