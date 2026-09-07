package syntaxfacts

import (
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/domain"
	"unicode/utf8"
)

// Decode validates report ownership and source ranges before downstream mapping.
func Decode(raw json.RawMessage, path, language, source string) (Report, error) {
	var report Report
	if len(raw) > 8<<20 || json.Unmarshal(raw, &report) != nil {
		return report, errors.New("INVALID_SYNTAX_REPORT")
	}
	if report.Path != path || report.Language != language {
		return report, errors.New("SYNTAX_SCOPE_MISMATCH")
	}
	return report, report.Validate(source)
}

// Validate rejects forged ranges, declaration cycles and invented semantic bindings.
// It permits missing or erroneous source constructs, recorded by diagnostics.
func (r Report) Validate(source string) error {
	invalid := errors.New("INVALID_SYNTAX_FACTS")
	if r.Schema != 1 || !r.Capabilities.Syntax || r.Capabilities.SemanticResolution || domain.ValidateRelativePath(r.Path, false) != nil || !utf8.ValidString(source) {
		return invalid
	}
	if len(r.Symbols)+len(r.References)+len(r.Imports)+len(r.Diagnostics) > 100000 {
		return invalid
	}
	span := func(start, end int) bool {
		return start >= 0 && end >= start && end <= len(source) && (start == len(source) || utf8.RuneStart(source[start])) && (end == len(source) || utf8.RuneStart(source[end]))
	}
	declarations := make(map[string]Declaration, len(r.Symbols))
	for _, d := range r.Symbols {
		if d.ID == "" || d.Kind == "" || !span(d.Start, d.End) {
			return invalid
		}
		if _, exists := declarations[d.ID]; exists {
			return invalid
		}
		declarations[d.ID] = d
		if len(d.Parameters) > 10000 {
			return invalid
		}
		for _, p := range d.Parameters {
			if !span(p.Start, p.End) || p.Start < d.Start || p.End > d.End {
				return invalid
			}
		}
	}
	for _, d := range r.Symbols {
		seen := map[string]bool{d.ID: true}
		current := d
		for current.ParentID != "" {
			parent, exists := declarations[current.ParentID]
			if !exists || seen[parent.ID] || current.Start < parent.Start || current.End > parent.End || len(seen) > 128 {
				return invalid
			}
			seen[parent.ID] = true
			current = parent
		}
	}
	for _, ref := range r.References {
		if ref.Resolution != "unresolved" || !span(ref.Start, ref.End) {
			return invalid
		}
	}
	for _, imp := range r.Imports {
		if imp.Resolution != "unresolved" || !span(imp.Start, imp.End) {
			return invalid
		}
	}
	for _, d := range r.Diagnostics {
		if !span(d.Start, d.End) {
			return invalid
		}
	}
	return nil
}
