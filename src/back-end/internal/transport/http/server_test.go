package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/authorization"
	"github.com/valio-projects/valio.code/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testAPI(t *testing.T) http.Handler {
	t.Helper()
	a, e := authorization.NewBootstrap(strings.Repeat("x", 32), nil)
	if e != nil {
		t.Fatal(e)
	}
	return (&Server{Auth: a, Catalog: &catalog.Service{Workspace: domain.Workspace{ID: "workspace-main"}}, Queries: &queries.Service{WorkspaceID: "workspace-main"}, Ingestion: &snapshots.Service{WorkspaceID: "workspace-main"}, Ready: func(context.Context) error { return nil }}).Handler()
}
func callAPI(h http.Handler, method, path, body, token, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://localhost:8090"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestHTTPAuthenticationStrictBodiesAndScope(t *testing.T) {
	h := testAPI(t)
	token := strings.Repeat("x", 32)
	tests := []struct {
		method, path, body, token, origin string
		status                            int
	}{{"GET", "/healthz", "", "", "", 200}, {"GET", "/readyz", "", "", "", 200}, {"GET", "/api/v1/workspace", "", "", "", 401}, {"GET", "/api/v1/workspace?workspaceId=other", "", token, "", 403}, {"GET", "/api/v1/workspace", "", token, "https://evil.example", 403}, {"POST", "/api/v1/search", `{"query":"x","scope":{"workspaceId":"other"}}`, token, "", 403}, {"POST", "/api/v1/search", `{"query":"(","scope":{"workspaceId":"workspace-main"}}`, token, "", 400}, {"POST", "/api/v1/search", `{"unknown":true}`, token, "", 400}, {"POST", "/api/v1/search", `{} {}`, token, "", 400}, {"GET", "/api/v1/types", "", token, "", 400}, {"POST", "/api/v1/session", `{"token":"` + token + `"}`, "", "", 403}}
	for _, tt := range tests {
		w := callAPI(h, tt.method, tt.path, tt.body, tt.token, tt.origin)
		if w.Code != tt.status {
			t.Errorf("%s %s status=%d want=%d body=%s", tt.method, tt.path, w.Code, tt.status, w.Body.String())
		}
	}
	huge := `{"query":"` + strings.Repeat("a", 16<<20) + `"}`
	if w := callAPI(h, "POST", "/api/v1/search", huge, token, ""); w.Code != 413 {
		t.Fatal("body limit", w.Code)
	}
}
func TestSessionCookieAndCrossOrigin(t *testing.T) {
	h := testAPI(t)
	token := strings.Repeat("x", 32)
	w := callAPI(h, "POST", "/api/v1/session", `{"token":"`+token+`"}`, "", "http://localhost:8090")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || strings.Contains(cookies[0].Value, token) {
		t.Fatal("insecure cookie")
	}
	r := httptest.NewRequest("GET", "http://localhost:8090/api/v1/workspace", nil)
	r.AddCookie(cookies[0])
	read := httptest.NewRecorder()
	h.ServeHTTP(read, r)
	if read.Code != 200 {
		t.Fatal(read.Code)
	}
	r = httptest.NewRequest("DELETE", "http://localhost:8090/api/v1/session", nil)
	r.AddCookie(cookies[0])
	bad := httptest.NewRecorder()
	h.ServeHTTP(bad, r)
	if bad.Code != 401 {
		t.Fatal("cookie mutation without Origin", bad.Code)
	}
}
func TestSecretSnapshotRejectedBeforeDatabase(t *testing.T) {
	h := testAPI(t)
	source := "password=\"actual-secret-value\""
	sum := sha256.Sum256([]byte(source))
	snapshot := agent.Snapshot{Version: 1, Files: []agent.File{{Path: ".env", Hash: hex.EncodeToString(sum[:]), Size: len(source), Content: source, Language: "text"}}}
	b, _ := json.Marshal(snapshot)
	sum = sha256.Sum256(b)
	snapshot.ID = hex.EncodeToString(sum[:])
	b, _ = json.Marshal(snapshots.IngestCommand{WorkspaceID: "workspace-main", RepositoryID: "repo", Snapshot: snapshot})
	w := callAPI(h, "POST", "/api/v1/ingestion", string(b), strings.Repeat("x", 32), "")
	if w.Code != 400 || strings.Contains(w.Body.String(), "actual-secret-value") {
		t.Fatal(w.Code, w.Body.String())
	}
}
