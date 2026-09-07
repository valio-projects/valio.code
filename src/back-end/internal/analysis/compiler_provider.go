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
	// ParseFile parses source using the supplied file set and parser mode.
	ParseFile(*token.FileSet, string, string, parser.Mode) (*ast.File, error)
}

// CompilerProvider checks a package and may report nonfatal diagnostics through onError.
type CompilerProvider interface {
	// Check type-checks files for path and fills info when successful or partially successful.
	Check(string, *token.FileSet, []*ast.File, *types.Info, func(error)) (*types.Package, error)
}

// StandardGoSyntax delegates parsing to the Go standard library.
type StandardGoSyntax struct{}

// ParseFile parses source with go/parser and preserves parser errors.
func (StandardGoSyntax) ParseFile(fs *token.FileSet, path, source string, mode parser.Mode) (*ast.File, error) {
	return parser.ParseFile(fs, path, source, mode)
}

// StandardGoCompiler invokes go/types with an optional custom import resolver.
type StandardGoCompiler struct{ Importer types.Importer }

// Check type-checks files, using importer.Default when Importer is nil.
func (p StandardGoCompiler) Check(path string, fs *token.FileSet, files []*ast.File, info *types.Info, onError func(error)) (*types.Package, error) {
	imports := p.Importer
	if imports == nil {
		imports = importer.Default()
	}
	config := types.Config{Importer: imports, Error: onError}
	return config.Check(path, fs, files, info)
}
