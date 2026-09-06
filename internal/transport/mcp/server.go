// Package mcp adapts the shared application queries to the official MCP SDK.
package mcp

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
)

// Server registers only implemented operations; unsupported graph tools are not advertised.
type Server struct{ sdk *mcp.Server }

// NewServer injects the same repository/query services used by the HTTP adapters.
func NewServer(catalog repositories.CatalogRepository, queryService queries.Service) (*Server, error) {
	sdk := mcp.NewServer(&mcp.Implementation{Name: "valio.code", Version: "0.1.0-dev"}, nil)
	registerCatalogTools(sdk, catalog)
	registerQueryTools(sdk, queryService)
	return &Server{sdk: sdk}, nil
}

// HTTPHandler returns Streamable HTTP. The API composition must wrap it in authorization.
func (s *Server) HTTPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s.sdk }, &mcp.StreamableHTTPOptions{Stateless: true})
}

// RunStdio serves the SDK's stdio transport until cancellation or client disconnect.
func (s *Server) RunStdio(ctx context.Context) error { return s.sdk.Run(ctx, &mcp.StdioTransport{}) }
