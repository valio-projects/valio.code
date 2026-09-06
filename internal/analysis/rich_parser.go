package analysis

import (
	"go/ast"
	"go/token"
)

func extractTypes(r *Report, fs *token.FileSet, file *ast.File, ids map[*ast.Ident]string, options Options, provider CompilerProvider) {
	r.Types = []TypeInfo{}
	r.Functions = []FunctionInfo{}
	r.ConstantGroups = []ConstantGroup{}
	r.TypeDiagnostics = []Diagnostic{}
	r.TypeCheckStatus = "not-run"
	info := (GoTypeChecker{Provider: provider}).Check(r, fs, file, ids, options)
	byName := parseTypes(r, fs, file, ids, info)
	parseMethods(r, fs, file, ids, info, byName)
	parseConstants(r, fs, file, ids, info, byName)
}
