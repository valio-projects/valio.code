package search

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func lex(s string) ([]token, error) {
	var out []token
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			i += n
			continue
		}
		if s[i] == '(' || s[i] == ')' {
			out = append(out, token{value: s[i : i+1], special: true})
			i++
			continue
		}
		t := token{}
		start := i
		for i < len(s) && ((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') || s[i] == '_') {
			i++
		}
		if i < len(s) && s[i] == ':' {
			t.field = strings.ToLower(s[start:i])
			i++
		} else {
			i = start
		}
		if i == len(s) {
			return nil, fmt.Errorf("missing value for filter %q", t.field)
		}
		if s[i] == '"' || s[i] == '\'' || s[i] == '/' {
			quote := s[i]
			t.quoted = true
			t.regex = quote == '/'
			i++
			var b strings.Builder
			closed := false
			for i < len(s) {
				c := s[i]
				i++
				if c == quote {
					closed = true
					break
				}
				if c == '\\' {
					if i == len(s) {
						return nil, fmt.Errorf("unfinished escape")
					}
					next := s[i]
					i++
					if t.regex {
						if next == quote {
							b.WriteByte(next)
						} else {
							b.WriteByte('\\')
							b.WriteByte(next)
						}
					} else {
						switch next {
						case '\\', '"', '\'':
							b.WriteByte(next)
						case 'n':
							b.WriteByte('\n')
						case 'r':
							b.WriteByte('\r')
						case 't':
							b.WriteByte('\t')
						default:
							return nil, fmt.Errorf("unsupported quoted escape \\%c", next)
						}
					}
				} else {
					b.WriteByte(c)
				}
			}
			if !closed {
				return nil, fmt.Errorf("unterminated quoted value")
			}
			t.value = b.String()
			if i < len(s) {
				r, _ := utf8.DecodeRuneInString(s[i:])
				if !unicode.IsSpace(r) && s[i] != ')' && s[i] != '(' {
					return nil, fmt.Errorf("expected separator after quoted value")
				}
			}
		} else {
			start = i
			for i < len(s) {
				r, n := utf8.DecodeRuneInString(s[i:])
				if unicode.IsSpace(r) || r == '(' || r == ')' {
					break
				}
				i += n
			}
			t.value = s[start:i]
		}
		if t.value == "" {
			return nil, fmt.Errorf("empty search value")
		}
		if t.field == "" && !t.quoted && (t.value == "AND" || t.value == "OR" || t.value == "NOT") {
			t.special = true
		}
		out = append(out, t)
	}
	return out, nil
}
