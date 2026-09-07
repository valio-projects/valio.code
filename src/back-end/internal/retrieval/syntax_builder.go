package retrieval

import (
	"context"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

// BuildSyntax builds source-grounded chunks from one validated non-Go syntax report.
// It does not infer compiler facts: headers and symbol representations use only written
// declaration metadata, while every searchable Text value remains an exact source range.
func (b Builder) BuildSyntax(ctx context.Context, source Source, report syntaxfacts.Report) ([]domainretrieval.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := b.validate(); err != nil {
		return nil, err
	}
	if err := validateSyntaxSource(source, report); err != nil {
		return nil, err
	}
	projects, err := normalizedProjects(source.ProjectIDs)
	if err != nil {
		return nil, err
	}
	source.ProjectIDs = projects
	declarations, children, err := validateSyntaxDeclarations(source.Content, report.Symbols)
	if err != nil {
		return nil, err
	}
	if len(declarations) == 0 {
		return b.syntaxFallback(ctx, source, 0, len(source.Content), "syntax-file", "", false)
	}
	chunks := make([]domainretrieval.Chunk, 0, len(declarations))
	position := 0
	for _, declaration := range declarations {
		if declaration.ParentID != "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if position < declaration.Start {
			fallback, err := b.syntaxFallback(ctx, source, position, declaration.Start, "syntax-gap", "", false)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, fallback...)
		}
		built, err := b.emitSyntaxDeclaration(ctx, source, declaration, children, "")
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, built...)
		position = declaration.End
	}
	if position < len(source.Content) {
		fallback, err := b.syntaxFallback(ctx, source, position, len(source.Content), "syntax-gap", "", false)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, fallback...)
	}
	return chunks, nil
}
