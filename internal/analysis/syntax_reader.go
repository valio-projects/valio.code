package analysis

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
)

func sourceText(fset *token.FileSet, n ast.Node) string {
	if n == nil {
		return ""
	}
	var b bytes.Buffer
	if err := format.Node(&b, fset, n); err != nil {
		return ""
	}
	return b.String()
}
func syntaxRange(fset *token.FileSet, n ast.Node) Range {
	if n == nil {
		return Range{}
	}
	return Range{fset.Position(n.Pos()).Offset, fset.Position(n.End()).Offset}
}
func visibility(name string) string {
	if ast.IsExported(name) {
		return "exported"
	}
	return "package"
}
func outlineID(path, kind string, offset int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d", path, kind, offset)))
	return hex.EncodeToString(sum[:16])
}
