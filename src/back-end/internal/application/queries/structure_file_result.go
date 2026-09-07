package queries

import "github.com/valio-projects/valio.code/internal/structuregraph"

// StructureFileResult is a query-selected subset of one source graph.
type StructureFileResult struct {
	FileID   string                `json:"fileId"`   // FileID identifies immutable source.
	Path     string                `json:"path"`     // Path is repository-relative source spelling.
	Language string                `json:"language"` // Language selects the grammar.
	Nodes    []structuregraph.Node `json:"nodes"`    // Nodes carries source-backed declarations and observations.
	Edges    []structuregraph.Edge `json:"edges"`    // Edges has only included endpoints or explicitly unresolved imports.
}
