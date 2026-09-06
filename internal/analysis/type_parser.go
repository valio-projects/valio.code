package analysis

import (
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strconv"
)

func parseTypes(r *Report, fs *token.FileSet, file *ast.File, ids map[*ast.Ident]string, info *gotypes.Info) map[string]int {
	byName := map[string]int{}
	// Only package type declarations populate this package catalog; nested
	// local type declarations remain in the syntax outline with distinct IDs.
	for _, decl := range file.Decls {
		g, ok := decl.(*ast.GenDecl)
		if !ok || g.Tok != token.TYPE {
			continue
		}
		for _, s := range g.Specs {
			ts := s.(*ast.TypeSpec)
			typ := typeExpression(fs, ts.Type, info)
			alias := ts.Assign.IsValid()
			kind := typ.Kind
			if alias {
				kind = "alias"
			}
			t := TypeInfo{ID: ids[ts.Name], Name: ts.Name.Name, QualifiedName: file.Name.Name + "." + ts.Name.Name, Kind: kind, Alias: alias, Exported: ast.IsExported(ts.Name.Name), Visibility: visibility(ts.Name.Name), UnderlyingType: typ, TypeParameters: parameters(fs, ts.TypeParams, info), Fields: []FieldInfo{}, Methods: []FunctionInfo{}, EmbeddedTypes: []TypeExpression{}, Constants: []ConstantInfo{}, Range: syntaxRange(fs, ts), NameRange: syntaxRange(fs, ts.Name), Layout: LayoutInfo{Status: "unknown", Reason: "no compiler target/build profile supplied; AST does not establish ABI layout"}, Evidence: "go-ast-syntax"}
			if info != nil {
				if obj := info.Defs[ts.Name]; obj != nil && obj.Type().Underlying() != gotypes.Typ[gotypes.Invalid] {
					t.ResolvedUnderlyingType = gotypes.TypeString(obj.Type().Underlying(), nil)
				}
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				for _, f := range st.Fields.List {
					names := f.Names
					embedded := len(names) == 0
					if embedded {
						name := baseIdentifier(f.Type)
						if name == nil {
							continue
						}
						names = []*ast.Ident{name}
					}
					for _, name := range names {
						field := FieldInfo{ID: ids[name], Name: name.Name, Type: typeExpression(fs, f.Type, info), Exported: ast.IsExported(name.Name), Visibility: visibility(name.Name), Embedded: embedded, Range: syntaxRange(fs, f), NameRange: syntaxRange(fs, name)}
						if embedded {
							field.ID = outlineID(r.Path, "embedded-field", field.Range.Start)
						}
						if f.Tag != nil {
							field.TagLiteral = f.Tag.Value
							field.Tag, _ = strconv.Unquote(f.Tag.Value)
						}
						t.Fields = append(t.Fields, field)
					}
				}
			}
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				for _, f := range iface.Methods.List {
					if ft, ok := f.Type.(*ast.FuncType); ok && len(f.Names) > 0 {
						for _, name := range f.Names {
							method := functionInfo(fs, &ast.FuncDecl{Name: name, Type: ft}, ids, info)
							method.Range = syntaxRange(fs, f)
							t.Methods = append(t.Methods, method)
						}
					} else {
						t.EmbeddedTypes = append(t.EmbeddedTypes, typeExpression(fs, f.Type, info))
					}
				}
			}
			byName[t.Name] = len(r.Types)
			r.Types = append(r.Types, t)
		}
	}

	return byName
}
