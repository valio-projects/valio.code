package retrieval

import (
	"context"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

func (b Builder) emitSyntaxDeclaration(ctx context.Context, source Source, declaration syntaxfacts.Declaration, children map[string][]syntaxfacts.Declaration, parentID string) ([]domainretrieval.Chunk, error) {
	kind := syntaxChunkKind(declaration.Kind)
	if declaration.End-declaration.Start <= b.maxBytes() {
		parent := b.syntaxChunk(source, declaration.Start, declaration.End, kind, parentID, false, false, declaration)
		result := []domainretrieval.Chunk{parent}
		for _, child := range children[declaration.ID] {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			built, err := b.emitSyntaxDeclaration(ctx, source, child, children, parent.ID)
			if err != nil {
				return nil, err
			}
			result = append(result, built...)
		}
		return result, nil
	}
	parent := b.syntaxChunk(source, declaration.Start, declaration.End, kind, parentID, true, false, declaration)
	parent.Text = ""
	parent.Representations = syntaxRepresentations(parent, declaration)
	result := []domainretrieval.Chunk{parent}
	position := declaration.Start
	for _, child := range children[declaration.ID] {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if position < child.Start {
			fallback, err := b.syntaxFallback(ctx, source, position, child.Start, "syntax-gap", parent.ID, true)
			if err != nil {
				return nil, err
			}
			result = append(result, fallback...)
		}
		built, err := b.emitSyntaxDeclaration(ctx, source, child, children, parent.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, built...)
		position = child.End
	}
	if position < declaration.End {
		fallback, err := b.syntaxFallback(ctx, source, position, declaration.End, "syntax-gap", parent.ID, true)
		if err != nil {
			return nil, err
		}
		result = append(result, fallback...)
	}
	return result, nil
}

func (b Builder) syntaxFallback(ctx context.Context, source Source, start, end int, label, parent string, split bool) ([]domainretrieval.Chunk, error) {
	if start == end {
		return nil, nil
	}
	if end-start <= b.maxBytes() {
		return []domainretrieval.Chunk{b.chunk(source, start, end, domainretrieval.ChunkFallback, parent, split, true, label, "", "")}, nil
	}
	return b.buildFallback(ctx, source, start, end, label, parent, true)
}
