package analysis

import (
	"go/ast"
	"go/constant"
	"go/token"
	gotypes "go/types"
)

func parseConstants(r *Report, fs *token.FileSet, file *ast.File, ids map[*ast.Ident]string, info *gotypes.Info, byName map[string]int) {
	for _, decl := range file.Decls {
		g, ok := decl.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		group := ConstantGroup{ID: outlineID(r.Path, "constant-group", fs.Position(g.Pos()).Offset), Kind: "go-const-group", Range: syntaxRange(fs, g), Members: []ConstantInfo{}}
		var inheritedType ast.Expr
		var inheritedValues []ast.Expr
		for _, spec := range g.Specs {
			v := spec.(*ast.ValueSpec)
			typ, values := v.Type, v.Values
			inherited := len(values) == 0
			if inherited {
				typ = inheritedType
				values = inheritedValues
			} else {
				inheritedType = typ
				inheritedValues = values
			}
			for i, name := range v.Names {
				c := ConstantInfo{ID: ids[name], GroupID: group.ID, Name: name.Name, Exported: ast.IsExported(name.Name), DeclaredType: sourceText(fs, typ), InheritedExpression: inherited, Resolution: "unresolved", Range: syntaxRange(fs, name), References: []Occurrence{}}
				if i < len(values) {
					c.Expression = sourceText(fs, values[i])
				}
				typeName := baseName(typ)
				if info != nil {
					if obj, ok := info.Defs[name].(*gotypes.Const); ok {
						c.ResolvedType = gotypes.TypeString(obj.Type(), nil)
						if obj.Val().Kind() != constant.Unknown {
							c.Resolution = "go-types"
							c.Value = obj.Val().ExactString()
						}
						if named, ok := gotypes.Unalias(obj.Type()).(*gotypes.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Name() == file.Name.Name {
							typeName = named.Obj().Name()
						}
						for _, o := range r.Occurrences {
							if o.Role == "use" && o.Resolution == "go-types" && o.SymbolID == c.ID {
								c.References = append(c.References, o)
							}
						}
					}
				}
				group.Members = append(group.Members, c)
				if ti, ok := byName[typeName]; ok {
					r.Types[ti].Constants = append(r.Types[ti].Constants, c)
				}
			}
		}
		r.ConstantGroups = append(r.ConstantGroups, group)
	}
}
