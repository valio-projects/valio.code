package analysis

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// SyntaxProvider and CompilerProvider isolate compiler libraries from report
// extraction. Projects may supply a module-aware importer without changing AST
// traversal or claiming a custom semantic model.
type SyntaxProvider interface {
	ParseFile(*token.FileSet, string, string, parser.Mode) (*ast.File, error)
}
type CompilerProvider interface {
	Check(string, *token.FileSet, []*ast.File, *types.Info, func(error)) (*types.Package, error)
}

type StandardGoSyntax struct{}

func (StandardGoSyntax) ParseFile(fs *token.FileSet, path, source string, mode parser.Mode) (*ast.File, error) {
	return parser.ParseFile(fs, path, source, mode)
}

type StandardGoCompiler struct{ Importer types.Importer }

func (p StandardGoCompiler) Check(path string, fs *token.FileSet, files []*ast.File, info *types.Info, onError func(error)) (*types.Package, error) {
	imports := p.Importer
	if imports == nil {
		imports = importer.Default()
	}
	config := types.Config{Importer: imports, Error: onError}
	return config.Check(path, fs, files, info)
}
