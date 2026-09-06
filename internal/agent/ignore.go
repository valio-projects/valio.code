package agent

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// .valioignore supplements Git ignores (and can also exclude tracked files).
// Patterns support # comments, ! re-inclusion, *, ?, **, a leading / anchor,
// and a trailing / for directories. Last matching rule wins. Character classes
// and backslash escapes are intentionally not supported; invalid rules fail.
type ignoreRule struct {
	pattern *regexp.Regexp
	include bool
}

func readIgnore(root string) ([]ignoreRule, error) {
	p := filepath.Join(root, ".valioignore")
	info, e := os.Lstat(p)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil || !info.Mode().IsRegular() {
		return nil, errors.New("invalid .valioignore file")
	}
	rootFS, e := os.OpenRoot(root)
	if e != nil {
		return nil, errors.New("cannot open repository root")
	}
	defer rootFS.Close()
	b, e := (FilesystemReader{Root: rootFS}).Read(".valioignore", 64*1024)
	if e != nil {
		return nil, errors.New("cannot read .valioignore")
	}
	return parseIgnore(string(b))
}
func parseIgnore(data string) ([]ignoreRule, error) {
	rules := []ignoreRule{}
	for _, line := range strings.Split(data, "\n") {
		p := strings.TrimSpace(line)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		include := strings.HasPrefix(p, "!")
		if include {
			p = strings.TrimPrefix(p, "!")
		}
		if p == "" || strings.ContainsAny(p, "\\[]\x00") {
			return nil, errors.New("unsupported .valioignore pattern")
		}
		anchored := strings.HasPrefix(p, "/")
		p = strings.TrimPrefix(p, "/")
		dir := strings.HasSuffix(p, "/")
		p = strings.TrimSuffix(p, "/")
		prefix := "(?:^|/)"
		if anchored || strings.Contains(p, "/") {
			prefix = "^"
		}
		var out strings.Builder
		out.WriteString(prefix)
		for i := 0; i < len(p); i++ {
			switch p[i] {
			case '*':
				if i+1 < len(p) && p[i+1] == '*' {
					i++
					if i+1 < len(p) && p[i+1] == '/' {
						i++
						out.WriteString("(?:.*/)?")
					} else {
						out.WriteString(".*")
					}
				} else {
					out.WriteString("[^/]*")
				}
			case '?':
				out.WriteString("[^/]")
			default:
				out.WriteString(regexp.QuoteMeta(string(p[i])))
			}
		}
		if dir {
			out.WriteString("/.*$")
		} else {
			out.WriteString("(?:/.*)?$")
		}
		re, e := regexp.Compile(out.String())
		if e != nil {
			return nil, errors.New("invalid .valioignore pattern")
		}
		rules = append(rules, ignoreRule{re, include})
	}
	return rules, nil
}
func ignored(path string, rules []ignoreRule) bool {
	result := false
	for _, r := range rules {
		if r.pattern.MatchString(path) {
			result = !r.include
		}
	}
	return result
}
