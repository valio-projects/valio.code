package mcp

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/application/queries"
)

func registerIntelligenceTools(server *mcp.Server, service queries.Service) {
	mcp.AddTool(server, &mcp.Tool{Name: "retrieval_search", Description: "Rank immutable source chunks using lexical BM25, symbol, structural, semantic or hybrid retrieval; return channel evidence and missing capabilities."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.RetrievalQuery) (*mcp.CallToolResult, Output, error) {
		v, e := service.Retrieve(ctx, input)
		return nil, Output{Data: v}, e
	})
	mcp.AddTool(server, &mcp.Tool{Name: "context_get", Description: "Build bounded parent-child context from one immutable source chunk, with citations and omitted items."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.ContextQuery) (*mcp.CallToolResult, Output, error) {
		v, e := service.Context(ctx, input)
		return nil, Output{Data: v}, e
	})
	for _, tool := range []struct {
		name string
		mode queries.GraphMode
	}{{"symbol_search", queries.GraphSymbols}, {"symbol_references", queries.GraphReferences}, {"graph_callers", queries.GraphCallers}, {"graph_callees", queries.GraphCallees}, {"graph_neighbors", queries.GraphNeighbors}, {"graph_paths", queries.GraphPaths}, {"symbol_reads", queries.GraphReads}, {"symbol_writes", queries.GraphWrites}} {
		mode := tool.mode
		mcp.AddTool(server, &mcp.Tool{Name: tool.name, Description: "Query bounded Go graph evidence in one pinned workspace/project view; unresolved semantic boundaries and truncation are explicit."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.GraphQuery) (*mcp.CallToolResult, Output, error) {
			input.Mode = mode
			v, e := service.Graph(ctx, input)
			return nil, Output{Data: v}, e
		})
	}
	mcp.AddTool(server, &mcp.Tool{Name: "syntax_query", Description: "Inspect written C/C++/C#/Java/JS/TS types, fields, methods, parameters and enum members in a pinned source view; semantic links remain unresolved."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.SyntaxQuery) (*mcp.CallToolResult, Output, error) {
		v, e := service.Syntax(ctx, input)
		return nil, Output{Data: v}, e
	})
	mcp.AddTool(server, &mcp.Tool{Name: "embedding_index", Description: "Index a bounded resumable page of policy-approved persisted representations with an operator-configured model profile."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.EmbeddingCommand) (*mcp.CallToolResult, Output, error) {
		v, e := service.IndexEmbeddings(ctx, input)
		return nil, Output{Data: v}, e
	})
	mcp.AddTool(server, &mcp.Tool{Name: "ai_probe", Description: "Check a configured embedding endpoint with synthetic text and verify vector dimension; never changes model installation."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.ProviderProbe) (*mcp.CallToolResult, Output, error) {
		v, e := service.ProbeProvider(ctx, input)
		return nil, Output{Data: v}, e
	})
}
