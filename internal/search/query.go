package search

import (
	"fmt"
)

type node struct {
	op           string
	left, right  *node
	field, value string
	regex        bool
}
type Query struct{ root *node }
type QueryParser struct{ MaxQueryBytes int }
type token struct {
	value   string
	special bool
	quoted  bool
	regex   bool
	field   string
}
type parser struct {
	tokens []token
	at     int
}

// Parse accepts AND, OR, NOT, parentheses, implicit AND, quoted values, and
// /RE2 expressions/. Precedence is NOT, then AND, then OR. A NOT evaluates
// against a whole file, not individual lines or candidate blocks.
func Parse(input string) (*Query, error) {
	return (QueryParser{}).Parse(input)
}

// Parse validates input under the configured byte limit and returns an opaque query.
func (p QueryParser) Parse(input string) (*Query, error) {
	limit := p.MaxQueryBytes
	if limit <= 0 {
		limit = 64 * 1024
	}
	if len(input) > limit {
		return nil, fmt.Errorf("query exceeds %d-byte limit", limit)
	}
	tokens, err := lex(input)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("query is empty")
	}
	grammar := parser{tokens: tokens}
	root, err := grammar.or()
	if err != nil {
		return nil, err
	}
	if grammar.at != len(tokens) {
		return nil, fmt.Errorf("unexpected token %q", tokens[grammar.at].value)
	}
	return &Query{root: root}, nil
}

func (p *parser) is(s string) bool {
	return p.at < len(p.tokens) && p.tokens[p.at].special && p.tokens[p.at].value == s
}
func (p *parser) or() (*node, error) {
	n, e := p.and()
	if e != nil {
		return nil, e
	}
	for p.is("OR") {
		p.at++
		r, e := p.and()
		if e != nil {
			return nil, e
		}
		n = &node{op: "OR", left: n, right: r}
	}
	return n, nil
}
func (p *parser) and() (*node, error) {
	n, e := p.unary()
	if e != nil {
		return nil, e
	}
	for p.at < len(p.tokens) && !p.is("OR") && !p.is(")") {
		if p.is("AND") {
			p.at++
		}
		r, e := p.unary()
		if e != nil {
			return nil, e
		}
		n = &node{op: "AND", left: n, right: r}
	}
	return n, nil
}
func (p *parser) unary() (*node, error) {
	if p.at == len(p.tokens) {
		return nil, fmt.Errorf("expected expression")
	}
	if p.is("NOT") {
		p.at++
		n, e := p.unary()
		return &node{op: "NOT", left: n}, e
	}
	if p.is("(") {
		p.at++
		n, e := p.or()
		if e != nil {
			return nil, e
		}
		if !p.is(")") {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		p.at++
		return n, nil
	}
	t := p.tokens[p.at]
	p.at++
	if t.special {
		return nil, fmt.Errorf("unexpected operator %q", t.value)
	}
	switch t.field {
	case "", "project", "repo", "file", "lang", "path", "content", "symbol", "kind", "test", "generated":
	default:
		return nil, fmt.Errorf("unsupported filter %q", t.field)
	}
	if t.field == "test" || t.field == "generated" {
		if t.regex || (t.value != "true" && t.value != "false") {
			return nil, fmt.Errorf("%s expects true or false", t.field)
		}
	}
	return &node{op: "term", field: t.field, value: t.value, regex: t.regex}, nil
}
