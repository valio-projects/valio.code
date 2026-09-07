package httpapi

import (
	"context"
	"net/http"
)

// registerIntelligence exposes tested application operations after their core
// implementations. No frontend state or provider credential enters this layer.
func (s *Server) registerIntelligence(mux *http.ServeMux) {
	bindJSON(mux, "POST /api/v1/retrieval/search", s.Queries.Retrieve)
	bindJSON(mux, "POST /api/v1/graph/query", s.Queries.Graph)
	bindJSON(mux, "POST /api/v1/context", s.Queries.Context)
	bindJSON(mux, "POST /api/v1/embeddings/index", s.Queries.IndexEmbeddings)
	bindJSON(mux, "POST /api/v1/ai/probe", s.Queries.ProbeProvider)
	bindJSON(mux, "POST /api/v1/syntax/query", s.Queries.Syntax)
	bindJSON(mux, "POST /api/v1/structure/graph", s.Queries.Structure)
}

func bindJSON[Input, Output any](mux *http.ServeMux, pattern string, handler func(context.Context, Input) (Output, error)) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var input Input
		if !decode(w, r, &input) {
			return
		}
		value, err := handler(r.Context(), input)
		respond(w, value, err)
	})
}
