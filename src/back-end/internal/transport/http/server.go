package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/authorization"
	"net/http"
	"time"
)

// Server composes authenticated administration, ingestion, query and MCP routes.
type Server struct {
	// Auth authenticates the local bootstrap principal and signed browser sessions.
	Auth *authorization.Bootstrap
	// Catalog handles project definitions and repository registration.
	Catalog *catalog.Service
	// Queries resolves pinned source, search and type reads.
	Queries *queries.Service
	// Ingestion validates and atomically publishes sanitized source uploads.
	Ingestion *snapshots.Service
	// Ready supplies the typed ready value at this boundary.
	Ready func(context.Context) error
	// MCP mounts the MCP endpoint inside the same authorization boundary.
	MCP http.Handler
}

// Handler returns the complete router wrapped by authorization, request IDs and body/time budgets.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerSession(mux)
	s.registerAdmin(mux)
	s.registerRead(mux)
	if s.MCP != nil {
		mux.Handle("/mcp", s.MCP)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := make([]byte, 16)
		_, _ = rand.Read(id)
		w.Header().Set("X-Request-ID", hex.EncodeToString(id))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == "/healthz" {
			if r.Method != "GET" {
				failure(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
				return
			}
			write(w, 200, map[string]string{"status": "ok"})
			return
		}
		if r.URL.Path == "/readyz" {
			if r.Method != "GET" {
				failure(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if s.Ready == nil || s.Ready(ctx) != nil {
				failure(w, 503, "NOT_READY")
				return
			}
			write(w, 200, map[string]string{"status": "ready"})
			return
		}
		if r.URL.Path != "/api/v1/session" || r.Method != "POST" {
			if !s.Auth.Authenticate(r) {
				failure(w, 401, "UNAUTHORIZED")
				return
			}
		}
		if !s.Auth.AllowedOrigin(r, false) {
			failure(w, 403, "FORBIDDEN_ORIGIN")
			return
		}
		if workspace := r.URL.Query().Get("workspaceId"); workspace != "" && workspace != string(s.Catalog.Workspace.ID) {
			failure(w, 403, "WORKSPACE_FORBIDDEN")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
		mux.ServeHTTP(w, r)
	})
}
