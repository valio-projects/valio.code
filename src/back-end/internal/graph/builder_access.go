package graph

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
)

type writeAccess struct {
	identifier *ast.Ident
	object     types.Object
	resolution codegraph.Resolution
	pureWrite  bool
}

// addAccessFacts emits local reads and writes without claiming data-flow,
// alias, or control-flow semantics. Assignment targets are classified before
// ordinary identifier reads so x = value is not also represented as a read of x.
func (b *graphBuilder) addAccessFacts(item parsedSource, info *types.Info, objects map[types.Object]string, groupProjects []string) {
	pureWrites := map[*ast.Ident]bool{}
	ast.Inspect(item.file, func(node ast.Node) bool {
		if b.truncated || b.ctx.Err() != nil {
			return false
		}
		switch value := node.(type) {
		case *ast.GenDecl:
			if value.Tok == token.VAR {
				b.addDeclarationInitializations(item.source, value, info, objects, groupProjects)
			}
		case *ast.AssignStmt:
			b.addAssignmentWrites(item.source, value, info, objects, groupProjects, pureWrites)
		case *ast.IncDecStmt:
			b.addWriteAccess(item.source, accessFor(value.X, info), objects, groupProjects)
		}
		return true
	})
	ast.Inspect(item.file, func(node ast.Node) bool {
		if b.truncated || b.ctx.Err() != nil {
			return false
		}
		identifier, ok := node.(*ast.Ident)
		if !ok || pureWrites[identifier] || identifier.Name == "_" {
			return true
		}
		b.addRead(item.source, identifier, info.Uses[identifier], objects, groupProjects)
		return true
	})
}

func (b *graphBuilder) addDeclarationInitializations(source Source, declaration *ast.GenDecl, info *types.Info, objects map[types.Object]string, projects []string) {
	for _, spec := range declaration.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok || len(value.Values) == 0 {
			continue
		}
		for _, identifier := range value.Names {
			object := info.Defs[identifier]
			if _, writable := object.(*types.Var); !writable || objects[object] == "" {
				continue
			}
			rangeValue := b.rangeOf(source, identifier)
			definitionID := nodeID("definition", source.ID, rangeValue.Start, rangeValue.End, objectKind(object)+"\x00"+identifier.Name)
			b.addEdge(codegraph.RelationWrites, definitionID, objects[object], codegraph.ResolutionExact, compatibleProjects(source.ProjectIDs, projects), evidence("go/types", &rangeValue))
			b.complete.ExactWrites++
		}
	}
}

func (b *graphBuilder) addAssignmentWrites(source Source, assignment *ast.AssignStmt, info *types.Info, objects map[types.Object]string, projects []string, pureWrites map[*ast.Ident]bool) {
	compound := assignment.Tok != token.ASSIGN && assignment.Tok != token.DEFINE
	for _, expression := range assignment.Lhs {
		access := accessFor(expression, info)
		if access.identifier == nil {
			continue
		}
		if assignment.Tok == token.DEFINE && info.Defs[access.identifier] != nil {
			if len(assignment.Rhs) > 0 {
				b.addInitializationWrite(source, access.identifier, info.Defs[access.identifier], objects, projects)
			}
			continue
		}
		b.addWriteAccess(source, access, objects, projects)
		if !compound && access.pureWrite {
			pureWrites[access.identifier] = true
		}
	}
}

func (b *graphBuilder) addInitializationWrite(source Source, identifier *ast.Ident, object types.Object, objects map[types.Object]string, projects []string) {
	if _, writable := object.(*types.Var); !writable || objects[object] == "" {
		return
	}
	rangeValue := b.rangeOf(source, identifier)
	definitionID := nodeID("definition", source.ID, rangeValue.Start, rangeValue.End, objectKind(object)+"\x00"+identifier.Name)
	b.addEdge(codegraph.RelationWrites, definitionID, objects[object], codegraph.ResolutionExact, compatibleProjects(source.ProjectIDs, projects), evidence("go/types", &rangeValue))
	b.complete.ExactWrites++
}

func (b *graphBuilder) addRead(source Source, identifier *ast.Ident, object types.Object, objects map[types.Object]string, projects []string) {
	if !readableObject(object) {
		return
	}
	rangeValue := b.rangeOf(source, identifier)
	referenceID := referenceID(source, rangeValue, identifier.Name)
	target := objects[object]
	if target != "" {
		b.addEdge(codegraph.RelationReads, referenceID, target, codegraph.ResolutionExact, compatibleProjects(source.ProjectIDs, projects), evidence("go/types", &rangeValue))
		b.complete.ExactReads++
		return
	}
	b.addEdge(codegraph.RelationReads, referenceID, "", codegraph.ResolutionUnresolved, source.ProjectIDs, evidence("go/types", &rangeValue))
	b.complete.UnresolvedReads++
}

func (b *graphBuilder) addWriteAccess(source Source, access writeAccess, objects map[types.Object]string, projects []string) {
	if access.identifier == nil {
		return
	}
	rangeValue := b.rangeOf(source, access.identifier)
	referenceID := referenceID(source, rangeValue, access.identifier.Name)
	target := objects[access.object]
	resolution := access.resolution
	if target == "" {
		resolution = codegraph.ResolutionUnresolved
	}
	if resolution == codegraph.ResolutionUnresolved {
		target = ""
	}
	b.addEdge(codegraph.RelationWrites, referenceID, target, resolution, compatibleProjects(source.ProjectIDs, projects), evidence("go/types", &rangeValue))
	switch resolution {
	case codegraph.ResolutionExact:
		b.complete.ExactWrites++
	case codegraph.ResolutionCandidate:
		b.complete.CandidateWrites++
	default:
		b.complete.UnresolvedWrites++
	}
}

func accessFor(expression ast.Expr, info *types.Info) writeAccess {
	switch value := expression.(type) {
	case *ast.Ident:
		return writeAccess{identifier: value, object: info.Uses[value], resolution: codegraph.ResolutionExact, pureWrite: true}
	case *ast.SelectorExpr:
		object := info.Uses[value.Sel]
		if selection := info.Selections[value]; selection != nil {
			object = selection.Obj()
		}
		return writeAccess{identifier: value.Sel, object: object, resolution: codegraph.ResolutionExact, pureWrite: true}
	case *ast.IndexExpr:
		access := accessFor(value.X, info)
		access.resolution = codegraph.ResolutionCandidate
		access.pureWrite = false
		return access
	case *ast.IndexListExpr:
		access := accessFor(value.X, info)
		access.resolution = codegraph.ResolutionCandidate
		access.pureWrite = false
		return access
	case *ast.StarExpr:
		access := accessFor(value.X, info)
		access.resolution = codegraph.ResolutionUnresolved
		access.pureWrite = false
		return access
	default:
		return writeAccess{resolution: codegraph.ResolutionUnresolved}
	}
}

func readableObject(object types.Object) bool {
	switch object.(type) {
	case *types.Var, *types.Const:
		return true
	default:
		return false
	}
}

func referenceID(source Source, rangeValue codegraph.ByteRange, name string) string {
	return nodeID("reference", source.ID, rangeValue.Start, rangeValue.End, name)
}
