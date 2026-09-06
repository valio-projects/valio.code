package analysis

import (
	"go/ast"
	"go/token"
	gotypes "go/types"
)

func parameters(fs *token.FileSet, list *ast.FieldList, info *gotypes.Info) []ParameterInfo {
	out := []ParameterInfo{}
	if list == nil {
		return out
	}
	for _, f := range list.List {
		typ := typeExpression(fs, f.Type, info)
		_, variadic := f.Type.(*ast.Ellipsis)
		if len(f.Names) == 0 {
			out = append(out, ParameterInfo{Type: typ, Variadic: variadic, Range: syntaxRange(fs, f)})
		}
		for _, n := range f.Names {
			out = append(out, ParameterInfo{Name: n.Name, Type: typ, Variadic: variadic, Range: syntaxRange(fs, n)})
		}
	}
	return out
}
func functionInfo(fs *token.FileSet, f *ast.FuncDecl, ids map[*ast.Ident]string, info *gotypes.Info) FunctionInfo {
	signature := *f
	signature.Body = nil
	signature.Doc = nil
	out := FunctionInfo{ID: ids[f.Name], Name: f.Name.Name, Signature: sourceText(fs, &signature), Exported: ast.IsExported(f.Name.Name), Visibility: visibility(f.Name.Name), TypeParameters: parameters(fs, f.Type.TypeParams, info), Parameters: parameters(fs, f.Type.Params, info), Returns: parameters(fs, f.Type.Results, info), Range: syntaxRange(fs, f), NameRange: syntaxRange(fs, f.Name), Evidence: "go-ast-syntax"}
	if recv := parameters(fs, f.Recv, info); len(recv) > 0 {
		out.Receiver = &recv[0]
	}
	return out
}
func baseName(e ast.Expr) string {
	if id := baseIdentifier(e); id != nil {
		return id.Name
	}
	return ""
}
func baseIdentifier(e ast.Expr) *ast.Ident {
	switch x := e.(type) {
	case *ast.Ident:
		return x
	case *ast.SelectorExpr:
		return x.Sel
	case *ast.StarExpr:
		return baseIdentifier(x.X)
	case *ast.IndexExpr:
		return baseIdentifier(x.X)
	case *ast.IndexListExpr:
		return baseIdentifier(x.X)
	case *ast.ParenExpr:
		return baseIdentifier(x.X)
	}
	return nil
}

func parseMethods(r *Report, fs *token.FileSet, file *ast.File, ids map[*ast.Ident]string, info *gotypes.Info, byName map[string]int) {
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok {
			fn := functionInfo(fs, f, ids, info)
			if f.Recv == nil {
				r.Functions = append(r.Functions, fn)
			} else if len(f.Recv.List) > 0 {
				if index, ok := byName[baseName(f.Recv.List[0].Type)]; ok {
					r.Types[index].Methods = append(r.Types[index].Methods, fn)
				} else {
					r.Functions = append(r.Functions, fn)
				}
			}
		}
	}

}
