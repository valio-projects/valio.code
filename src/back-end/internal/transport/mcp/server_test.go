package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
	"github.com/valio-projects/valio.code/internal/projects"
)

type catalogFixture struct{ repositories.CatalogRepository }

func (catalogFixture) Projects(context.Context) ([]projects.Definition, error) {
	return []projects.Definition{{Project: domain.Project{ID: "service-users", Name: "Users"}}}, nil
}

func TestStreamableToolDiscoveryAndInvocation(t *testing.T) {
	server, err := NewServer(catalogFixture{}, queries.Service{})
	if err != nil {
		t.Fatal(err)
	}
	host := httptest.NewServer(server.HTTPHandler())
	defer host.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: host.URL, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 19 {
		t.Fatalf("advertised unimplemented tools: %d", len(tools.Tools))
	}
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "project_list", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("tool call failed: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("unexpected content: %d", len(result.Content))
	}
	content, ok := result.Content[0].(*sdk.TextContent)
	if !ok || !strings.Contains(content.Text, "service-users") {
		t.Fatalf("project result missing")
	}
}

func TestBridgeEndpointAndRedirectProtection(t *testing.T) {
	for _, endpoint := range []string{"http://example.com", "https://user:pass@example.com", "https://example.com?token=x"} {
		if _, err := NewBridge(endpoint, strings.Repeat("a", 32)); err == nil {
			t.Errorf("accepted unsafe endpoint %s", endpoint)
		}
	}
	reached := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer source.Close()
	bridge, err := NewBridge(source.URL, strings.Repeat("a", 32))
	if err != nil {
		t.Fatal(err)
	}
	response, err := bridge.client.Get(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusFound || reached {
		t.Fatal("credential-bearing redirect followed")
	}
}
