package mcp

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
)

// ProjectInput selects an existing stable project identity in the authenticated workspace.
type ProjectInput struct {
	// ID is the stable project identity in the authenticated workspace.
	ID string `json:"id" jsonschema:"Stable project ID"`
}

func registerCatalogTools(server *mcp.Server, catalog repositories.CatalogRepository) {
	mcp.AddTool(server, &mcp.Tool{Name: "project_list", Description: "List projects in the authenticated workspace."}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, Output, error) {
		rows, err := catalog.Projects(ctx)
		return nil, Output{Data: rows}, err
	})
	mcp.AddTool(server, &mcp.Tool{Name: "project_get", Description: "Read a project definition and its repository source roots."}, func(ctx context.Context, _ *mcp.CallToolRequest, input ProjectInput) (*mcp.CallToolResult, Output, error) {
		row, err := catalog.Project(ctx, input.ID)
		return nil, Output{Data: row}, err
	})
}
