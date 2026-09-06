package mcp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Bridge forwards stdio tools to an authenticated Streamable HTTP endpoint.
// It holds only an API token and never receives database credentials.
type Bridge struct {
	endpoint string
	client   *http.Client
}

// NewBridge accepts HTTPS or explicit loopback HTTP and refuses redirects so a
// server cannot redirect the bearer token to a different origin.
func NewBridge(endpoint, token string) (*Bridge, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || len(token) < 32 {
		return nil, errors.New("invalid MCP endpoint or token")
	}
	loopback := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil, errors.New("MCP endpoint requires HTTPS except loopback")
	}
	return &Bridge{endpoint: strings.TrimRight(endpoint, "/") + "/mcp", client: &http.Client{Timeout: 60 * time.Second, Transport: &bearerTransport{token: token, base: http.DefaultTransport}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

// Run exposes only the tools currently advertised by the remote server. All
// input validation, authorization and version resolution remain on that server.
func (b *Bridge) Run(ctx context.Context) error {
	client := mcp.NewClient(&mcp.Implementation{Name: "valio-mcp", Version: "0.1.0-dev"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: b.endpoint, HTTPClient: b.client, DisableStandaloneSSE: true}, nil)
	if err != nil {
		return errors.New("MCP server connection failed")
	}
	defer session.Close()
	server := mcp.NewServer(&mcp.Implementation{Name: "valio-mcp", Version: "0.1.0-dev"}, nil)
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return errors.New("MCP tool discovery failed")
		}
		name := tool.Name
		server.AddTool(tool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: req.Params.Arguments})
		})
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}
