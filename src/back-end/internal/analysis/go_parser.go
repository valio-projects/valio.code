// Package analysis extracts Go syntax evidence and optionally checks local
// objects with Go's type checker. Compiler calls, flow, and ABI remain unknown.
package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"sort"
	"strconv"
)

func (p Parser) Analyze(path, content string) Report {
	options := p.Options
	r := Report{Path: path, Symbols: []Symbol{}, Occurrences: []Occurrence{}, Imports: []Import{}, Diagnostics: []Diagnostic{}, Capabilities: Capabilities{SyntaxOutline: true, DefinitionRanges: true}}
	r.TypeCheckStatus = "not-run"
	r.Types = []TypeInfo{}
	r.Functions = []FunctionInfo{}
	r.ConstantGroups = []ConstantGroup{}
	r.TypeDiagnostics = []Diagnostic{}
	fset := token.NewFileSet()
	syntax := p.Syntax
	if syntax == nil {
		syntax = StandardGoSyntax{}
	}
	file, err := syntax.ParseFile(fset, path, content, parser.AllErrors|parser.ParseComments)
	r.ValidSyntax = err == nil
	if err != nil {
		if es, ok := err.(scanner.ErrorList); ok {
			for _, e := range es {
				r.Diagnostics = append(r.Diagnostics, Diagnostic{Message: e.Msg, Offset: e.Pos.Offset, Line: e.Pos.Line, Column: e.Pos.Column})
			}
		} else {
			r.Diagnostics = append(r.Diagnostics, Diagnostic{Message: err.Error()})
		}
	}
	if file == nil {
		return r
	}
	r.Package = file.Name.Name
	rangeOf := func(n ast.Node) Range {
		if n == nil {
			return Range{}
		}
		return Range{min(fset.Position(n.Pos()).Offset, len(content)), min(fset.Position(n.End()).Offset, len(content))}
	}
	definitions := map[*ast.Ident]string{}
	add := func(id *ast.Ident, kind string, decl ast.Node) {
		if id == nil || id.Name == "_" {
			return
		}
		if _, ok := definitions[id]; ok {
			return
		}
		nr := rangeOf(id)
		digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%s", path, nr.Start, kind, id.Name)))
		sid := hex.EncodeToString(digest[:16])
		definitions[id] = sid
		r.Symbols = append(r.Symbols, Symbol{ID: sid, Name: id.Name, Kind: kind, Range: nr, DeclarationRange: rangeOf(decl), Evidence: "go-ast-syntax"})
	}
	fields := func(list *ast.FieldList, kind string) {
		if list == nil {
			return
		}
		for _, f := range list.List {
			for _, id := range f.Names {
				add(id, kind, f)
			}
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncDecl:
			kind := "function"
			if n.Recv != nil {
				kind = "method"
			}
			add(n.Name, kind, n)
			fields(n.Recv, "receiver")
		case *ast.FuncType:
			fields(n.TypeParams, "type-parameter")
			fields(n.Params, "parameter")
			fields(n.Results, "result")
		case *ast.TypeSpec:
			add(n.Name, "type", n)
			fields(n.TypeParams, "type-parameter")
		case *ast.GenDecl:
			for _, s := range n.Specs {
				if v, ok := s.(*ast.ValueSpec); ok {
					kind := "variable"
					if n.Tok == token.CONST {
						kind = "constant"
					}
					for _, id := range v.Names {
						add(id, kind, v)
					}
				}
			}
		case *ast.StructType:
			fields(n.Fields, "field")
		case *ast.InterfaceType:
			fields(n.Methods, "interface-method")
		case *ast.ImportSpec:
			p, e := strconv.Unquote(n.Path.Value)
			if e != nil {
				p = n.Path.Value
			}
			im := Import{Path: p, Range: rangeOf(n)}
			if n.Name != nil {
				im.Alias = n.Name.Name
				if n.Name.Name != "." {
					add(n.Name, "import-alias", n)
				}
			}
			r.Imports = append(r.Imports, im)
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, x := range n.Lhs {
					if id, ok := x.(*ast.Ident); ok && id.Obj != nil && id.Obj.Decl == n {
						add(id, "variable", n)
					}
				}
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				for _, x := range []ast.Expr{n.Key, n.Value} {
					if id, ok := x.(*ast.Ident); ok {
						add(id, "variable", n)
					}
				}
			}
		case *ast.LabeledStmt:
			add(n.Label, "label", n)
		}
		return true
	})
	ast.Inspect(file, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			role, resolution := "use", "unresolved"
			sid := definitions[id]
			if sid != "" {
				role = "definition"
				resolution = "syntax-definition"
			}
			if id == file.Name {
				role = "package-clause"
				resolution = "syntax"
			}
			if id.Name == "_" {
				role = "blank"
				resolution = "not-applicable"
			}
			r.Occurrences = append(r.Occurrences, Occurrence{Name: id.Name, Range: rangeOf(id), Role: role, SymbolID: sid, Resolution: resolution})
		}
		return true
	})
	sort.Slice(r.Symbols, func(i, j int) bool { return r.Symbols[i].Range.Start < r.Symbols[j].Range.Start })
	sort.Slice(r.Occurrences, func(i, j int) bool { return r.Occurrences[i].Range.Start < r.Occurrences[j].Range.Start })
	extractTypes(&r, fset, file, definitions, options, p.Compiler)
	return r
}
