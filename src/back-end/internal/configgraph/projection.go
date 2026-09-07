// Package configgraph extracts configuration key locations. It never returns
// configuration values, source lines, or hashes of configuration values.
package configgraph

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Key struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Line   int    `json:"line,omitempty"`
	Status string `json:"status"`
}
type Projection struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Keys []Key  `json:"keys"`
}

var envKey = regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=`)

func Kind(path string) string {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))
	if strings.HasPrefix(base, ".env") || strings.HasSuffix(base, ".env") {
		return "env"
	}
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml", ".ini", ".conf", ".config", ".properties", ".tfvars", ".hcl", ".xml":
		return "opaque"
	}
	return ""
}

func Extract(path, kind string, data []byte) Projection {
	p := Projection{Path: path, Kind: kind, Keys: []Key{}}
	if kind == "json" {
		var value any
		if json.Unmarshal(data, &value) == nil {
			walkJSON(path, "", value, &p.Keys)
		}
	} else if kind == "env" {
		// A quoted value can contain newlines and text resembling assignments.
		// Track quotes so those lines never become bogus keys or leaked values.
		var quote byte
		for i, line := range bytes.Split(data, []byte{'\n'}) {
			s := strings.TrimSpace(string(line))
			if quote != 0 {
				quote = closingQuote(s, quote)
				continue
			}
			if strings.HasPrefix(s, "#") {
				continue
			}
			m := envKey.FindStringSubmatch(s)
			if m == nil {
				continue
			}
			p.Keys = append(p.Keys, Key{Name: m[1], Path: path, Line: i + 1, Status: "value-redacted"})
			value := strings.TrimSpace(s[len(m[0]):])
			if len(value) > 0 && (value[0] == '\'' || value[0] == '"') {
				quote = closingQuote(value[1:], value[0])
			}
		}
	} else if kind == "yaml" {
		// YAML is intentionally excluded rather than regex-parsed: quoted keys,
		// aliases, tags and block scalars can turn apparent keys into secret values.
	}
	sort.Slice(p.Keys, func(i, j int) bool {
		if p.Keys[i].Name == p.Keys[j].Name {
			return p.Keys[i].Line < p.Keys[j].Line
		}
		return p.Keys[i].Name < p.Keys[j].Name
	})
	return p
}
func closingQuote(s string, q byte) byte {
	escaped := false
	for i := 0; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' && q == '"' {
			escaped = true
			continue
		}
		if s[i] == q {
			return 0
		}
	}
	return q
}

var safeKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,127}$`)

func walkJSON(path, prefix string, v any, keys *[]Key) {
	switch x := v.(type) {
	case map[string]any:
		for k, c := range x {
			if !safeKey.MatchString(k) {
				continue
			}
			name := k
			if prefix != "" {
				name = prefix + "." + k
			}
			*keys = append(*keys, Key{Name: name, Path: path, Status: "value-redacted"})
			walkJSON(path, name, c, keys)
		}
	case []any:
		for _, c := range x {
			walkJSON(path, prefix+"[]", c, keys)
		}
	}
}
