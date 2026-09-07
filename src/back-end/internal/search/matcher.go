package search

import (
	"fmt"
	"path"
	"regexp"
	"regexp/syntax"
	"strings"
	"unicode/utf8"
)

type compiled struct {
	n           *node
	left, right *compiled
	re          *regexp.Regexp
	mode        Mode
	value       string
	required    []string
	sensitive   bool
}

func compile(n *node, o Options) (*compiled, error) {
	c := &compiled{n: n, mode: o.Mode, value: n.value, sensitive: o.CaseSensitive}
	if c.mode == "" {
		c.mode = Substring
	}
	var e error
	if n.op != "term" {
		if n.left != nil {
			c.left, e = compile(n.left, o)
			if e != nil {
				return nil, e
			}
		}
		if n.right != nil {
			c.right, e = compile(n.right, o)
		}
		return c, e
	}
	if n.field != "" && n.field != "content" && n.field != "path" && n.field != "file" && n.field != "symbol" {
		c.mode = Exact
	}
	if n.regex {
		c.mode = Regex
	}
	if c.mode == Regex {
		pattern := n.value
		if !o.CaseSensitive {
			pattern = "(?i:" + pattern + ")"
		}
		c.re, e = regexp.Compile(pattern)
		if e != nil {
			return nil, fmt.Errorf("invalid regex: %w", e)
		}
		parsed, e := syntax.Parse(pattern, syntax.Perl)
		if e != nil {
			return nil, e
		}
		literal := requiredLiteral(parsed)
		c.required = trigrams(fold(literal))
	} else {
		c.required = trigrams(fold(n.value))
		if !o.CaseSensitive {
			c.value = fold(c.value)
		}
	}
	return c, nil
}

// requiredLiteral is deliberately conservative. Alternations, optional nodes,
// and character classes have no extracted requirement. A concat's mandatory
// child and a repetition with Min>0 are safe regardless of regex anchoring.
func requiredLiteral(r *syntax.Regexp) string {
	switch r.Op {
	case syntax.OpLiteral:
		return string(r.Rune)
	case syntax.OpCapture:
		return requiredLiteral(r.Sub[0])
	case syntax.OpPlus:
		return requiredLiteral(r.Sub[0])
	case syntax.OpRepeat:
		if r.Min > 0 {
			return requiredLiteral(r.Sub[0])
		}
	case syntax.OpConcat:
		best := ""
		for _, s := range r.Sub {
			v := requiredLiteral(s)
			if len([]rune(v)) > len([]rune(best)) {
				best = v
			}
		}
		return best
	}
	return ""
}
func (c *compiled) candidate(idx Index) bool {
	switch c.n.op {
	case "NOT":
		return true
	case "AND":
		return c.left.candidate(idx) && c.right.candidate(idx)
	case "OR":
		return c.left.candidate(idx) || c.right.candidate(idx)
	}
	if c.n.field != "" && c.n.field != "content" {
		return true
	}
	for _, t := range c.required {
		if _, ok := idx.Trigrams[t]; !ok {
			return false
		}
	}
	return true
}

func (c *compiled) matches(value string) rangeMatches {
	if c.mode == Regex {
		// Request one extra occurrence as a sentinel, so regex evaluation never
		// allocates an unbounded result slice for a single file field.
		hits := c.re.FindAllStringIndex(value, maxRangesPerMatch+1)
		out := rangeMatches{ranges: make([]Range, 0, min(len(hits), maxRangesPerMatch))}
		for _, h := range hits {
			out.add(Range{h[0], h[1]})
		}
		return out
	}
	text := value
	var starts, ends []int
	if !c.sensitive {
		var b strings.Builder
		for start, r := range value {
			_, size := utf8.DecodeRuneInString(value[start:])
			next := start + size
			v := string(foldRune(r))
			b.WriteString(v)
			for range len(v) {
				starts = append(starts, start)
				ends = append(ends, next)
			}
		}
		text = b.String()
	}
	if c.mode == Exact {
		if text == c.value {
			return rangeMatches{ranges: []Range{{0, len(value)}}}
		}
		return rangeMatches{}
	}
	out := rangeMatches{}
	for at := 0; at <= len(text)-len(c.value); {
		p := strings.Index(text[at:], c.value)
		if p < 0 {
			break
		}
		start := at + p
		end := start + len(c.value)
		if c.sensitive {
			out.add(Range{start, end})
		} else {
			out.add(Range{starts[start], ends[end-1]})
		}
		if out.truncated {
			return out
		}
		at = end
	}
	return out
}
func (c *compiled) evaluate(f File, projects []string) (bool, rangeMatches) {
	switch c.n.op {
	case "NOT":
		ok, _ := c.left.evaluate(f, projects)
		// NOT is exact at the file level but deliberately has no positive spans.
		return !ok, rangeMatches{}
	case "AND":
		a, ar := c.left.evaluate(f, projects)
		if !a {
			return false, rangeMatches{}
		}
		b, br := c.right.evaluate(f, projects)
		if !b {
			return false, rangeMatches{}
		}
		return true, combineRanges(ar, br)
	case "OR":
		a, ar := c.left.evaluate(f, projects)
		b, br := c.right.evaluate(f, projects)
		if !a {
			ar = rangeMatches{}
		}
		if !b {
			br = rangeMatches{}
		}
		return a || b, combineRanges(ar, br)
	}
	var values []string
	switch c.n.field {
	case "", "content":
		r := c.matches(f.Content)
		return len(r.ranges) > 0, r
	case "path":
		values = []string{f.Path}
	case "file":
		values = []string{path.Base(strings.ReplaceAll(f.Path, "\\", "/"))}
	case "lang":
		values = []string{f.Language}
	case "repo":
		values = []string{f.RepositoryID}
	case "project":
		values = projects
	case "test":
		return f.Test == (c.n.value == "true"), rangeMatches{}
	case "generated":
		return f.Generated == (c.n.value == "true"), rangeMatches{}
	case "symbol", "kind":
		ranges := rangeMatches{}
		ok := false
		matchingSymbols := 0
		for _, s := range f.Symbols {
			v := s.Name
			if c.n.field == "kind" {
				v = s.Kind
			}
			if len(c.matches(v).ranges) > 0 {
				ok = true
				matchingSymbols++
				if matchingSymbols > maxRangesPerMatch {
					ranges.truncated = true
					break
				}
				if s.Start >= 0 && s.End >= s.Start && s.End <= len(f.Content) {
					ranges.add(Range{s.Start, s.End})
				}
			}
		}
		return ok, ranges
	}
	for _, v := range values {
		if len(c.matches(v).ranges) > 0 {
			return true, rangeMatches{}
		}
	}
	return false, rangeMatches{}
}
