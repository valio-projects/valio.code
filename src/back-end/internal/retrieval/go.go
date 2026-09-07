package retrieval

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

type goUnit struct {
	start, end  int
	kind        domainretrieval.ChunkKind
	label       string
	declaration string
	doc         string
	body        *ast.BlockStmt
}

func (b Builder) buildGo(ctx context.Context, source Source) ([]domainretrieval.Chunk, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, source.Path, source.Content, parser.ParseComments|parser.AllErrors)
	if err != nil || file == nil {
		return b.buildFallback(ctx, source, 0, len(source.Content), "go-fallback", "", false)
	}
	units := []goUnit{}
	for _, declaration := range file.Decls {
		switch node := declaration.(type) {
		case *ast.FuncDecl:
			kind := domainretrieval.ChunkFunction
			label := "function"
			if node.Recv != nil {
				kind, label = domainretrieval.ChunkMethod, "method"
			}
			units = append(units, goUnit{start: offset(fset, node.Pos(), len(source.Content)), end: offset(fset, node.End(), len(source.Content)), kind: kind, label: label, declaration: node.Name.Name, doc: commentText(node.Doc), body: node.Body})
		case *ast.GenDecl:
			if node.Tok != token.TYPE {
				continue
			}
			for _, spec := range node.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				units = append(units, goUnit{start: offset(fset, typeSpec.Pos(), len(source.Content)), end: offset(fset, typeSpec.End(), len(source.Content)), kind: domainretrieval.ChunkType, label: "type", declaration: typeSpec.Name.Name, doc: commentText(node.Doc)})
			}
		}
	}
	sort.Slice(units, func(i, j int) bool {
		if units[i].start != units[j].start {
			return units[i].start < units[j].start
		}
		return units[i].end < units[j].end
	})
	chunks := []domainretrieval.Chunk{}
	for _, unit := range units {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		built, err := b.emitGoUnit(ctx, source, fset, unit)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, built...)
	}
	if len(chunks) == 0 && source.Content != "" {
		return b.buildFallback(ctx, source, 0, len(source.Content), "go-file", "", false)
	}
	return chunks, nil
}

func (b Builder) emitGoUnit(ctx context.Context, source Source, fset *token.FileSet, unit goUnit) ([]domainretrieval.Chunk, error) {
	if unit.end-unit.start <= b.maxBytes() {
		return []domainretrieval.Chunk{b.chunk(source, unit.start, unit.end, unit.kind, "", false, false, unit.label, unit.declaration, unit.doc)}, nil
	}
	parent := b.chunk(source, unit.start, unit.end, unit.kind, "", true, false, unit.label, unit.declaration, unit.doc)
	parent.Text = ""
	parent.Representations = representations(parent, unit.declaration, unit.doc)
	result := []domainretrieval.Chunk{parent}
	if unit.body != nil && len(unit.body.List) > 0 {
		for _, statement := range unit.body.List {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			start, end := offset(fset, statement.Pos(), len(source.Content)), offset(fset, statement.End(), len(source.Content))
			if end-start <= b.maxBytes() {
				result = append(result, b.chunk(source, start, end, domainretrieval.ChunkStatement, parent.ID, true, false, "statement", unit.declaration, ""))
				continue
			}
			fallback, err := b.buildFallback(ctx, source, start, end, "statement", parent.ID, true)
			if err != nil {
				return nil, err
			}
			result = append(result, fallback...)
		}
		return result, nil
	}
	fallback, err := b.buildFallback(ctx, source, unit.start, unit.end, unit.label, parent.ID, true)
	if err != nil {
		return nil, err
	}
	return append(result, fallback...), nil
}

func offset(fset *token.FileSet, position token.Pos, max int) int {
	value := fset.Position(position).Offset
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}

func commentText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return strings.TrimSpace(group.Text())
}
