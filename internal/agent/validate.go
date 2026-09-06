package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/configgraph"
)

var projectedKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.\[\]-]*$`)
var credentialURL = regexp.MustCompile(`(?i)(?:https?|postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis)://[^\s/@:]+:[^\s/@]+@`)
var diagnosticMessages = map[string]string{
	"unsafe-path":             "path leaves repository root",
	"unreadable":              "file missing or unreadable",
	"symlink-excluded":        "symlink paths are excluded",
	"nonregular-excluded":     "non-regular file excluded",
	"sensitive-excluded":      "sensitive artifact excluded before reading",
	"oversize-excluded":       "file exceeds capture size limit",
	"binary-excluded":         "binary or non-UTF-8 content excluded",
	"config-projected":        "configuration values excluded; only safe key metadata retained",
	"secret-pattern-excluded": "content containing a potential credential excluded before hashing",
}

// ValidateSnapshot rechecks the safe payload policy before hashes, spool writes,
// or transport. Identity alone is insufficient: a forged payload can rehash itself.
func ValidateSnapshot(s Snapshot) error {
	invalid := func() error { return errors.New("invalid or unsafe snapshot payload") }
	if s.Version != 1 {
		return invalid()
	}
	if !validateRepositoryMetadata(s.Repository) {
		return invalid()
	}
	seen := map[string]bool{}
	for _, f := range s.Files {
		if !safeRelativePath(f.Path) || seen[f.Path] || sensitiveName(f.Path) || configgraph.Kind(f.Path) != "" || strings.HasPrefix(f.Path, ".git/") || strings.HasPrefix(f.Path, ".valio/") {
			return invalid()
		}
		seen[f.Path] = true
		if len(f.Content) > int(DefaultMaxFileBytes) || f.Size != len(f.Content) || !utf8.ValidString(f.Content) || strings.ContainsRune(f.Content, 0) || suspiciousContent([]byte(f.Content)) || f.Language != language(f.Path) {
			return invalid()
		}
		sum := sha256.Sum256([]byte(f.Content))
		if f.Hash != hex.EncodeToString(sum[:]) {
			return invalid()
		}
	}
	for _, p := range s.Config {
		if !safeRelativePath(p.Path) || seen[p.Path] || sensitiveName(p.Path) || p.Kind == "" || configgraph.Kind(p.Path) != p.Kind {
			return invalid()
		}
		seen[p.Path] = true
		if p.Kind != "env" && p.Kind != "json" && len(p.Keys) > 0 {
			return invalid()
		}
		for _, k := range p.Keys {
			if k.Path != p.Path || len(k.Name) > 4096 || !projectedKey.MatchString(k.Name) || k.Status != "value-redacted" || k.Line < 0 || (p.Kind == "env" && k.Line == 0) || (p.Kind == "json" && k.Line != 0) {
				return invalid()
			}
		}
	}
	for _, d := range s.Diagnostics {
		if d.Path != "" && !safeRelativePath(d.Path) {
			return invalid()
		}
		expected, ok := diagnosticMessages[d.Code]
		if !ok {
			return invalid()
		}
		if d.Message != expected && !(d.Code == "unreadable" && d.Message == "file cannot be read within size limit") {
			return invalid()
		}
	}
	if !validSnapshot(s) {
		return invalid()
	}
	return nil
}
func safeRelativePath(p string) bool {
	if p == "" || len(p) > 32768 || strings.ContainsAny(p, "\\\x00\r\n") || containsControl(p) || strings.Contains(p, ":") || strings.HasPrefix(p, "/") || path.Clean(p) != p || p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return false
	}
	return utf8.ValidString(p)
}
