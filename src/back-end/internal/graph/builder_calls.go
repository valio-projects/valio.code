package graph

import (
	"go/ast"
	"go/types"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
)

// addCallableContainment records the lexical function-to-call-site relation
// needed to ask for a callable symbol's callees. It is parser evidence and
// does not infer a call target; RelationCalls remains go/types-only.
func (b *graphBuilder) addCallableContainment(item parsedSource, info *types.Info, objects map[types.Object]string, groupProjects []string) {
	for _, declaration := range item.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		ownerID := objects[info.Defs[function.Name]]
		if ownerID == "" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if b.truncated || b.ctx.Err() != nil {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			rangeValue := b.rangeOf(item.source, call)
			b.addEdge(codegraph.RelationContains, ownerID, callSiteID(item.source, rangeValue), codegraph.ResolutionExact, compatibleProjects(item.source.ProjectIDs, groupProjects), evidence("go/parser", &rangeValue))
			return true
		})
	}
}

func callSiteID(source Source, rangeValue codegraph.ByteRange) string {
	return nodeID("call", source.ID, rangeValue.Start, rangeValue.End, "")
}
