package analysis

import (
	"go/ast"
	"go/token"
	gotypes "go/types"
)

type GoTypeChecker struct{ Provider CompilerProvider }

func (checker GoTypeChecker) Check(r *Report, fs *token.FileSet, file *ast.File, ids map[*ast.Ident]string, options Options) *gotypes.Info {
	var info *gotypes.Info
	if options.CheckTypes && r.ValidSyntax {
		info = &gotypes.Info{Types: map[ast.Expr]gotypes.TypeAndValue{}, Defs: map[*ast.Ident]gotypes.Object{}, Uses: map[*ast.Ident]gotypes.Object{}, Selections: map[*ast.SelectorExpr]*gotypes.Selection{}}
		onError := func(err error) {
			d := Diagnostic{Message: err.Error()}
			if e, ok := err.(gotypes.Error); ok {
				p := fs.Position(e.Pos)
				d = Diagnostic{Message: e.Msg, Offset: p.Offset, Line: p.Line, Column: p.Column}
			}
			r.TypeDiagnostics = append(r.TypeDiagnostics, d)
		}
		pkg := options.PackagePath
		if pkg == "" {
			pkg = file.Name.Name
		}
		provider := checker.Provider
		if provider == nil {
			provider = StandardGoCompiler{}
		}
		_, err := provider.Check(pkg, fs, []*ast.File{file}, info, onError)
		r.TypeCheckStatus = "complete"
		r.Capabilities.TypeChecked = err == nil
		if err != nil {
			r.TypeCheckStatus = "partial"
		}
		// Only a checker object mapped to a local syntax definition receives a
		// local symbol ID. Imported and missing objects are left unresolved.
		objects := map[gotypes.Object]string{}
		for id, obj := range info.Defs {
			if sid := ids[id]; obj != nil && sid != "" {
				objects[obj] = sid
			}
		}
		uses := map[int]gotypes.Object{}
		for id, obj := range info.Uses {
			uses[fs.Position(id.Pos()).Offset] = obj
		}
		for i := range r.Occurrences {
			o := &r.Occurrences[i]
			if o.Role != "use" {
				continue
			}
			if obj := uses[o.Range.Start]; obj != nil {
				if sid := objects[obj]; sid != "" {
					o.SymbolID = sid
					o.Resolution = "go-types"
					o.ObjectType = gotypes.TypeString(obj.Type(), nil)
				}
			}
		}
	}

	return info
}
