package structuregraph

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

// MaxNodes bounds facts emitted from one syntax helper report.
const MaxNodes = 200000

// MaxEdges bounds relations emitted from one syntax helper report.
const MaxEdges = 400000

// ErrBudget reports that Build stopped after reaching a local graph bound.
var ErrBudget = errors.New("structure graph budget exceeded")

type builder struct {
	ctx       context.Context
	source    Source
	report    syntaxfacts.Report
	nodes     map[string]Node
	edges     map[string]Edge
	truncated bool
}

// Build maps one validated syntax helper report into a bounded declaration
// graph. It preserves helper-reported lexical ownership and unresolved import
// and reference observations, but it never resolves a name, import, member, or
// call target. It returns a partial graph with ErrBudget at a local bound and
// returns ctx.Err() when cancellation is observed.
func Build(ctx context.Context, source Source, report syntaxfacts.Report) (Graph, error) {
	if err := validateInput(source, report); err != nil {
		return Graph{}, err
	}
	b := builder{ctx: ctx, source: copiedSource(source), report: report, nodes: map[string]Node{}, edges: map[string]Edge{}}
	if err := b.build(); err != nil {
		return b.graph(), err
	}
	graph := b.graph()
	if err := graph.Validate(); err != nil {
		return graph, err
	}
	if b.truncated {
		return graph, ErrBudget
	}
	return graph, nil
}

func validateInput(source Source, report syntaxfacts.Report) error {
	if source.ID == "" || source.RepositoryID == "" || source.Path == "" || !utf8.ValidString(source.Content) || report.Path != source.Path || report.Language == "" {
		return errors.New("invalid structure graph input")
	}
	projects := map[string]bool{}
	for _, projectID := range source.ProjectIDs {
		if projectID == "" || projects[projectID] {
			return errors.New("invalid structure graph source projects")
		}
		projects[projectID] = true
	}
	if err := report.Validate(source.Content); err != nil {
		return fmt.Errorf("invalid syntax facts: %w", err)
	}
	return nil
}

func copiedSource(source Source) Source {
	copy := source
	copy.ProjectIDs = sortedStrings(source.ProjectIDs)
	return copy
}

func (b *builder) build() error {
	fileRange := b.rangeOf(0, len(b.source.Content))
	fileID := sourceNodeID(NodeFile, b.source.ID, "file", fileRange.Start, fileRange.End)
	b.addNode(Node{ID: fileID, Kind: NodeFile, Name: b.source.Path, FileID: b.source.ID, RepositoryID: b.source.RepositoryID, ProjectIDs: b.source.ProjectIDs, Range: fileRange, Evidence: b.evidence(fileRange)})
	declarations := map[string]string{}
	for _, declaration := range b.report.Symbols {
		if err := b.ctx.Err(); err != nil {
			return err
		}
		if b.truncated {
			return nil
		}
		rangeValue := b.rangeOf(declaration.Start, declaration.End)
		id := sourceNodeID(NodeDeclaration, b.source.ID, declaration.ID, rangeValue.Start, rangeValue.End)
		declarations[declaration.ID] = id
		b.addNode(Node{ID: id, Kind: NodeDeclaration, Name: declaration.Name, DeclarationKind: declaration.Kind, FileID: b.source.ID, RepositoryID: b.source.RepositoryID, ProjectIDs: b.source.ProjectIDs, Range: rangeValue, Type: declaration.Type, UnderlyingType: declaration.UnderlyingType, EnumValue: declaration.EnumValue, Visibility: declaration.Visibility, Modifiers: copiedStrings(declaration.Modifiers), Attributes: copiedStrings(declaration.Attributes), Evidence: b.evidence(rangeValue)})
		b.addEdge(RelationContains, fileID, id, ResolutionExact, rangeValue)
	}
	for _, declaration := range b.report.Symbols {
		if err := b.ctx.Err(); err != nil {
			return err
		}
		if b.truncated {
			return nil
		}
		id := declarations[declaration.ID]
		rangeValue := b.rangeOf(declaration.Start, declaration.End)
		ownerID := fileID
		if declaration.ParentID != "" {
			ownerID = declarations[declaration.ParentID]
		}
		b.addEdge(RelationDeclares, ownerID, id, ResolutionExact, rangeValue)
		for ordinal, parameter := range declaration.Parameters {
			if err := b.ctx.Err(); err != nil {
				return err
			}
			if b.truncated {
				return nil
			}
			parameterRange := b.rangeOf(parameter.Start, parameter.End)
			parameterID := sourceNodeID(NodeParameter, b.source.ID, declaration.ID+"/parameter/"+strconv.Itoa(ordinal), parameterRange.Start, parameterRange.End)
			b.addNode(Node{ID: parameterID, Kind: NodeParameter, Name: parameter.Name, DeclarationKind: "parameter", FileID: b.source.ID, RepositoryID: b.source.RepositoryID, ProjectIDs: b.source.ProjectIDs, Range: parameterRange, Type: parameter.Type, Modifiers: copiedStrings(parameter.Modifiers), Attributes: copiedStrings(parameter.Attributes), Evidence: b.evidence(parameterRange)})
			b.addEdge(RelationContains, fileID, parameterID, ResolutionExact, parameterRange)
			b.addEdge(RelationHasParameter, id, parameterID, ResolutionExact, parameterRange)
		}
	}
	for ordinal, imported := range b.report.Imports {
		if err := b.ctx.Err(); err != nil {
			return err
		}
		if b.truncated {
			return nil
		}
		rangeValue := b.rangeOf(imported.Start, imported.End)
		id := sourceNodeID(NodeImport, b.source.ID, imported.Path+"/import/"+strconv.Itoa(ordinal), rangeValue.Start, rangeValue.End)
		b.addNode(Node{ID: id, Kind: NodeImport, Name: imported.Path, DeclarationKind: imported.Kind, FileID: b.source.ID, RepositoryID: b.source.RepositoryID, ProjectIDs: b.source.ProjectIDs, Range: rangeValue, Evidence: b.evidence(rangeValue)})
		b.addEdge(RelationContains, fileID, id, ResolutionExact, rangeValue)
		b.addEdge(RelationImports, id, "", ResolutionUnresolved, rangeValue)
	}
	for ordinal, reference := range b.report.References {
		if err := b.ctx.Err(); err != nil {
			return err
		}
		if b.truncated {
			return nil
		}
		rangeValue := b.rangeOf(reference.Start, reference.End)
		kind := NodeReference
		if reference.Kind == "call" {
			kind = NodeCall
		}
		id := sourceNodeID(kind, b.source.ID, reference.Kind+"/reference/"+strconv.Itoa(ordinal), rangeValue.Start, rangeValue.End)
		b.addNode(Node{ID: id, Kind: kind, Name: reference.Name, DeclarationKind: reference.Kind, FileID: b.source.ID, RepositoryID: b.source.RepositoryID, ProjectIDs: b.source.ProjectIDs, Range: rangeValue, Evidence: b.evidence(rangeValue)})
		b.addEdge(RelationContains, fileID, id, ResolutionExact, rangeValue)
	}
	return nil
}

func (b *builder) rangeOf(start, end int) ByteRange {
	return ByteRange{FileID: b.source.ID, Start: start, End: end}
}

func (b *builder) evidence(rangeValue ByteRange) Evidence {
	return Evidence{Producer: "syntaxfacts/" + b.report.Language, Range: &rangeValue}
}

func (b *builder) addNode(node Node) {
	if b.nodes[node.ID].ID != "" {
		return
	}
	if len(b.nodes) >= MaxNodes {
		b.truncated = true
		return
	}
	b.nodes[node.ID] = node
}

func (b *builder) addEdge(kind RelationKind, sourceID, targetID string, resolution Resolution, rangeValue ByteRange) {
	if b.truncated {
		return
	}
	id := relationID(kind, sourceID, targetID, resolution, rangeValue)
	if b.edges[id].ID != "" {
		return
	}
	if len(b.edges) >= MaxEdges {
		b.truncated = true
		return
	}
	b.edges[id] = Edge{ID: id, Kind: kind, SourceID: sourceID, TargetID: targetID, Resolution: resolution, ProjectIDs: b.source.ProjectIDs, Evidence: b.evidence(rangeValue)}
}

func (b *builder) graph() Graph {
	nodes := make([]Node, 0, len(b.nodes))
	for _, node := range b.nodes {
		nodes = append(nodes, node)
	}
	edges := make([]Edge, 0, len(b.edges))
	for _, edge := range b.edges {
		edges = append(edges, edge)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return Graph{Schema: SchemaVersion, Language: b.report.Language, ValidSyntax: b.report.ValidSyntax, Nodes: nodes, Edges: edges}
}

func copiedStrings(values []string) []string {
	return append([]string{}, values...)
}

func sortedStrings(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}
