package mcp

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/search"
)

type typeQueryFixture struct {
	repositories.SnapshotRepository
}

func (typeQueryFixture) Latest(context.Context) (snapshots.View, error) {
	return snapshots.View{ID: "view", WorkspaceID: "workspace", Projections: map[string]string{"retrieval_chunks": "ready", "code_graph": "partial"}}, nil
}
func (typeQueryFixture) Artifacts(context.Context, snapshots.View) ([]snapshots.Artifact, error) {
	return []snapshots.Artifact{}, nil
}

func (typeQueryFixture) Files(context.Context, snapshots.View) ([]search.File, error) {
	return []search.File{}, nil
}

// Omitted optional scope fields must have the same defaults in MCP and HTTP.
func TestTypeQueryAcceptsOptionalScopeDefaults(t *testing.T) {
	server, err := NewServer(catalogFixture{}, queries.Service{Store: typeQueryFixture{}, WorkspaceID: "workspace"})
	if err != nil {
		t.Fatal(err)
	}
	host := httptest.NewServer(server.HTTPHandler())
	defer host.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := sdk.NewClient(&sdk.Implementation{Name: "type-query-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: host.URL, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	for _, test := range []struct {
		name string
		args map[string]any
	}{
		{"type_query", map[string]any{"name": "User"}},
		{"code_search", map[string]any{"query": "User", "scope": map[string]any{"workspaceId": "workspace"}}},
		{"retrieval_search", map[string]any{"query": "User", "scope": map[string]any{"workspaceId": "workspace"}}},
		{"symbol_search", map[string]any{"name": "User", "scope": map[string]any{"workspaceId": "workspace"}}},
	} {
		result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: test.name, Arguments: test.args})
		if err != nil || result.IsError {
			t.Fatalf("%s optional query fields rejected: result=%+v error=%v", test.name, result, err)
		}
	}
}
