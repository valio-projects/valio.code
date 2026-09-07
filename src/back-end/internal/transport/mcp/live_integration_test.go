package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestLiveHTTPMCPAgreement reads only the explicitly selected smoke fixture.
// Enable it with VALIO_TEST_API_URL, VALIO_TEST_PROJECT_ID, VALIO_TEST_VIEW_ID
// and VALIO_TEST_API_TOKEN after running scripts/smoke.ps1 against Compose.
func TestLiveHTTPMCPAgreement(t *testing.T) {
	endpoint := os.Getenv("VALIO_TEST_API_URL")
	if endpoint == "" {
		t.Skip("requires an explicit running Compose API and pinned fixture")
	}
	project, view := os.Getenv("VALIO_TEST_PROJECT_ID"), os.Getenv("VALIO_TEST_VIEW_ID")
	if project == "" || view == "" {
		t.Fatal("live parity test requires a project and immutable view")
	}
	bridge, err := NewBridge(endpoint, os.Getenv("VALIO_TEST_API_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := sdk.NewClient(&sdk.Implementation{Name: "http-mcp-parity-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: bridge.endpoint, HTTPClient: bridge.client, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 19 {
		t.Fatalf("unexpected tool count: %d", len(listed.Tools))
	}
	scope := map[string]any{"workspaceId": "workspace-main", "viewId": view, "projectIds": []string{project}}
	for _, test := range []struct {
		name, path, method string
		arguments          map[string]any
	}{
		{"structure_graph", "/api/v1/structure/graph", http.MethodPost, map[string]any{"scope": scope, "name": "User", "depth": 3}},
		{"type_query", "/api/v1/types", http.MethodGet, map[string]any{"name": "User", "projectId": project, "viewId": view}},
		{"code_search", "/api/v1/search", http.MethodPost, map[string]any{"scope": scope, "query": "User"}},
		{"retrieval_search", "/api/v1/retrieval/search", http.MethodPost, map[string]any{"scope": scope, "query": "User"}},
		{"symbol_search", "/api/v1/graph/query", http.MethodPost, map[string]any{"scope": scope, "name": "User"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: test.name, Arguments: test.arguments})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError {
				t.Fatal("MCP rejected a request accepted by the HTTP contract")
			}
			raw, err := json.Marshal(result.StructuredContent)
			if err != nil {
				t.Fatal(err)
			}
			var wrapped struct {
				Data any `json:"data"`
			}
			if err := json.Unmarshal(raw, &wrapped); err != nil {
				t.Fatal(err)
			}
			if test.name == "symbol_search" {
				test.arguments["mode"] = "symbols"
			}
			body, err := json.Marshal(test.arguments)
			if err != nil {
				t.Fatal(err)
			}
			var reader io.Reader
			if test.method == http.MethodPost {
				reader = bytes.NewReader(body)
			}
			request, err := http.NewRequestWithContext(ctx, test.method, strings.TrimRight(endpoint, "/")+test.path, reader)
			if err != nil {
				t.Fatal(err)
			}
			if test.method == http.MethodGet {
				query := request.URL.Query()
				for key, value := range test.arguments {
					query.Set(key, value.(string))
				}
				request.URL.RawQuery = query.Encode()
			}
			request.Header.Set("Content-Type", "application/json")
			response, err := bridge.client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("HTTP returned %d", response.StatusCode)
			}
			var output any
			if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&output); err != nil {
				t.Fatal(err)
			}
			if wrapped.Data == nil || !reflect.DeepEqual(wrapped.Data, output) {
				t.Fatal("MCP and HTTP results differ within the same pinned scope")
			}
		})
	}
}
