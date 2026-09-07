package httpapi

import (
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/projects"
	"net/http"
)

func (s *Server) registerAdmin(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		var body projects.Definition
		if !decode(w, r, &body) {
			return
		}
		v, e := (catalog.ProjectCommandHandler{Service: s.Catalog}).Handle(r.Context(), catalog.SaveProjectCommand{Definition: body})
		respond(w, v, e)
	})
	mux.HandleFunc("PUT /api/v1/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body projects.Definition
		if !decode(w, r, &body) {
			return
		}
		if string(body.Project.ID) != r.PathValue("id") {
			failure(w, 400, "IDENTITY_MISMATCH")
			return
		}
		v, e := (catalog.ProjectCommandHandler{Service: s.Catalog}).Handle(r.Context(), catalog.SaveProjectCommand{Definition: body, Update: true})
		respond(w, v, e)
	})
	mux.HandleFunc("POST /api/v1/repositories", func(w http.ResponseWriter, r *http.Request) {
		var body domain.Repository
		if !decode(w, r, &body) {
			return
		}
		v, e := (catalog.RepositoryCommandHandler{Service: s.Catalog}).Handle(r.Context(), catalog.RegisterRepositoryCommand{Repository: body})
		respond(w, v, e)
	})
	mux.HandleFunc("POST /api/v1/ingestion", func(w http.ResponseWriter, r *http.Request) {
		var body snapshots.IngestCommand
		if !decode(w, r, &body) {
			return
		}
		v, e := (snapshots.CommandHandler{Service: s.Ingestion}).Handle(r.Context(), body)
		respond(w, v, e)
	})
}
