package snapshots

import (
	"context"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
	"github.com/valio-projects/valio.code/internal/retrieval"
	"github.com/valio-projects/valio.code/internal/structuregraph"
	"github.com/valio-projects/valio.code/internal/types"
)

// syntaxArtifactProcessor constructs related projections from one validated
// analyzer result. Consumers cannot accidentally observe different source parses.
type syntaxArtifactProcessor struct{ analyzer SyntaxAnalyzer }

// Process enriches one unpublished artifact; failure leaves the current view intact.
func (p syntaxArtifactProcessor) Process(ctx context.Context, view View, file FileRef, source string, artifact *Artifact) error {
	raw, err := p.analyzer.Analyze(ctx, file.Path, file.Language, source)
	if err != nil {
		return err
	}
	report, err := syntaxfacts.Decode(raw, file.Path, file.Language, source)
	if err != nil {
		return err
	}
	artifact.Syntax = raw
	chunks, err := (retrieval.Builder{}).BuildSyntax(ctx, retrieval.Source{ID: file.ID, RepositoryID: string(file.RepositoryID), Path: file.Path, Language: file.Language, Content: source, ProjectIDs: file.ProjectIDs}, report)
	if err != nil {
		return err
	}
	artifact.Chunks = chunks
	graph, err := structuregraph.Build(ctx, structuregraph.Source{ID: file.ID, RepositoryID: string(file.RepositoryID), Path: file.Path, Content: source, ProjectIDs: file.ProjectIDs}, report)
	if err != nil {
		return err
	}
	artifact.StructureGraph, err = json.Marshal(graph)
	if err != nil {
		return err
	}
	for _, project := range file.ProjectIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		scope := typeinfo.TypeScope{WorkspaceID: view.WorkspaceID, ProjectID: domain.ProjectID(project), BuildProfileID: "syntax-default", VersionID: view.ID}
		descriptors, err := types.FromSyntaxReport(scope, file.RepositoryID, source, report)
		if err != nil {
			return err
		}
		artifact.Types = append(artifact.Types, descriptors...)
	}
	return nil
}
