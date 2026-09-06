package httpapi

import (
	"github.com/valio-projects/valio.code/internal/application/queries"
	"net/http"
)

func (s *Server) registerRead(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspace", func(w http.ResponseWriter, r *http.Request) { write(w, 200, s.Catalog.Workspace) })
	mux.HandleFunc("GET /api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Catalog.Store.Projects(r.Context())
		respond(w, map[string]any{"items": v}, e)
	})
	mux.HandleFunc("GET /api/v1/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Catalog.Store.Project(r.Context(), r.PathValue("id"))
		respond(w, v, e)
	})
	mux.HandleFunc("GET /api/v1/repositories", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Catalog.Store.Repositories(r.Context())
		respond(w, map[string]any{"items": v}, e)
	})
	mux.HandleFunc("GET /api/v1/views", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Queries.View(r.Context(), "")
		respond(w, v, e)
	})
	mux.HandleFunc("GET /api/v1/views/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Queries.View(r.Context(), r.PathValue("id"))
		respond(w, v, e)
	})
	mux.HandleFunc("GET /api/v1/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Queries.File(r.Context(), r.URL.Query().Get("viewId"), r.PathValue("id"))
		respond(w, v, e)
	})
	mux.HandleFunc("POST /api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		var q queries.SearchQuery
		if !decode(w, r, &q) {
			return
		}
		v, e := (queries.SearchQueryHandler{Service: s.Queries}).Handle(r.Context(), q)
		respond(w, v, e)
	})
	mux.HandleFunc("GET /api/v1/types", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		v, e := (queries.TypeQueryHandler{Service: s.Queries}).Handle(r.Context(), queries.TypeQuery{Name: q.Get("name"), ViewID: q.Get("viewId"), ProjectID: q.Get("projectId"), BuildProfileID: q.Get("buildProfileId")})
		respond(w, v, e)
	})
	mux.HandleFunc("GET /api/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		features := []map[string]string{}
		for _, name := range []string{"source_text", "symbols", "types", "compiler", "references", "git_diff", "structural_fingerprint", "configuration_graph"} {
			status := "unsupported"
			if name == "source_text" {
				status = "ready"
			}
			if name == "symbols" || name == "types" {
				status = "partial"
			}
			features = append(features, map[string]string{"name": name, "status": status})
		}
		write(w, 200, map[string]any{"features": features, "limits": map[string]int{"maxBodyBytes": 16 << 20, "maxFiles": 10000, "maxScanBytes": 16 << 20}, "authentication": "local-bootstrap", "buildProfileId": "syntax-default"})
	})
}
