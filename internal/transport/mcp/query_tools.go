package mcp

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/application/queries"
)

// ViewInput pins a view; omitting ViewID resolves latest exactly once per request.
type ViewInput struct {
	// ViewID optionally pins an immutable view; absent resolves latest once.
	ViewID string `json:"viewId,omitempty"`
}

func registerQueryTools(server *mcp.Server, service queries.Service) {
	mcp.AddTool(server, &mcp.Tool{Name: "code_search", Description: "Verified substring, exact or RE2 search in one immutable project view. No semantic or graph inference."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.SearchQuery) (*mcp.CallToolResult, Output, error) {
		result, err := service.Search(ctx, input)
		return nil, Output{Data: result}, err
	})
	mcp.AddTool(server, &mcp.Tool{Name: "type_query", Description: "Find a type by simple or qualified name; inspect members, parameters, attributes, constants, evidence and unknown layout. Ambiguity is explicit."}, func(ctx context.Context, _ *mcp.CallToolRequest, input queries.TypeQuery) (*mcp.CallToolResult, Output, error) {
		result, err := service.Types(ctx, input)
		return nil, Output{Data: result}, err
	})
	mcp.AddTool(server, &mcp.Tool{Name: "index_status", Description: "Read the pinned source view and its actual projection completeness."}, func(ctx context.Context, _ *mcp.CallToolRequest, input ViewInput) (*mcp.CallToolResult, Output, error) {
		result, err := service.View(ctx, input.ViewID)
		return nil, Output{Data: result}, err
	})
}
