package analysis

import (
	"go/ast"
	"go/constant"
	"go/token"
	gotypes "go/types"
)

func typeExpression(fs *token.FileSet, e ast.Expr, info *gotypes.Info) TypeExpression {
	t := TypeExpression{Text: sourceText(fs, e), Kind: "named", Evidence: "go-ast-syntax"}
	switch n := e.(type) {
	case *ast.StarExpr:
		t.Kind = "pointer"
		el := typeExpression(fs, n.X, info)
		t.Element = &el
	case *ast.ArrayType:
		if n.Len == nil {
			t.Kind = "slice"
		} else {
			t.Kind = "array"
			d := ArrayDimension{Expression: sourceText(fs, n.Len), Resolution: "unresolved"}
			if info != nil {
				if tv, ok := info.Types[n.Len]; ok && tv.Value != nil && tv.Value.Kind() == constant.Int {
					if length, ok := constant.Int64Val(tv.Value); ok && length >= 0 {
						d.Length = &length
						d.Resolution = "go-types"
					}
				}
			}
			t.Dimensions = append(t.Dimensions, d)
		}
		el := typeExpression(fs, n.Elt, info)
		t.Element = &el
		if n.Len != nil && el.Kind == "array" {
			t.Dimensions = append(t.Dimensions, el.Dimensions...)
		}
	case *ast.MapType:
		t.Kind = "map"
	case *ast.ChanType:
		t.Kind = "channel"
		el := typeExpression(fs, n.Value, info)
		t.Element = &el
	case *ast.StructType:
		t.Kind = "struct"
	case *ast.InterfaceType:
		t.Kind = "interface"
	case *ast.FuncType:
		t.Kind = "function"
	case *ast.Ellipsis:
		t.Kind = "variadic"
		el := typeExpression(fs, n.Elt, info)
		t.Element = &el
	case *ast.IndexExpr, *ast.IndexListExpr:
		t.Kind = "instantiation"
	case *ast.ParenExpr:
		return typeExpression(fs, n.X, info)
	}
	return t
}
