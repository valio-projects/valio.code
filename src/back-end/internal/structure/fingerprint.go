// Package structure computes heuristic syntax fingerprints, not equivalence
// proofs. It retains identifier spelling, operators, and control-node kinds.
package structure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

const Scheme = "go-ast-shape-v1"

type Result struct {
	Hash        string `json:"hash"`
	Scheme      string `json:"scheme"`
	Evidence    string `json:"evidence"`
	Equivalence string `json:"equivalence"`
}

// Fingerprint ignores comments, positions, and literal values, but preserves
// literal token kinds, operators, identifiers, and control structure. It does
// not normalize binding names: resolving shadowing requires separate evidence.
// Matching hashes are structural candidates only; e.g. return 1 and return 2
// intentionally match and may have different observable behavior.
func Fingerprint(source string) (Result, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "input.go", source, parser.AllErrors)
	if err != nil {
		return Result{}, fmt.Errorf("cannot fingerprint invalid Go syntax: %w", err)
	}
	var b strings.Builder
	encode(&b, reflect.ValueOf(f))
	sum := sha256.Sum256([]byte(b.String()))
	return Result{Hash: hex.EncodeToString(sum[:]), Scheme: Scheme, Evidence: "syntax-shape", Equivalence: "unknown"}, nil
}

var positionType = reflect.TypeOf(token.Pos(0))
var literalType = reflect.TypeOf(ast.BasicLit{})

func encode(b *strings.Builder, v reflect.Value) {
	if !v.IsValid() {
		b.WriteString("nil;")
		return
	}
	if v.Type() == positionType {
		return
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			b.WriteString("nil;")
		} else {
			encode(b, v.Elem())
		}
	case reflect.Struct:
		b.WriteString(v.Type().String())
		b.WriteByte('{')
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			switch f.Name {
			case "Obj", "Scope", "Unresolved", "Doc", "Comment", "Comments", "GoVersion":
				continue
			}
			if f.Type == positionType {
				continue
			}
			if v.Type() == literalType && f.Name == "Value" {
				continue
			}
			b.WriteString(f.Name)
			b.WriteByte('=')
			encode(b, v.Field(i))
		}
		b.WriteByte('}')
	case reflect.Slice:
		b.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			encode(b, v.Index(i))
		}
		b.WriteByte(']')
	case reflect.String:
		b.WriteString(strconv.Quote(v.String()))
		b.WriteByte(';')
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(v.Int(), 10))
		b.WriteByte(';')
	case reflect.Bool:
		b.WriteString(strconv.FormatBool(v.Bool()))
		b.WriteByte(';')
	default:
		panic("unsupported AST fingerprint field: " + v.Type().String())
	}
}
