package httpapi

import "net/http"

func (s *Server) registerSession(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/session", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			// Token holds the bootstrap credential only in server process memory.
			Token string `json:"token"`
		}
		if !decode(w, r, &body) {
			return
		}
		if !s.Auth.AllowedOrigin(r, true) {
			failure(w, 403, "FORBIDDEN_ORIGIN")
			return
		}
		if !s.Auth.CheckToken(body.Token) {
			failure(w, 401, "UNAUTHORIZED")
			return
		}
		if s.Auth.SetSession(w, r) != nil {
			failure(w, 500, "INTERNAL_ERROR")
			return
		}
		write(w, 200, map[string]string{"status": "authenticated"})
	})
	mux.HandleFunc("DELETE /api/v1/session", func(w http.ResponseWriter, r *http.Request) {
		s.Auth.ClearSession(w, r)
		write(w, 200, map[string]string{"status": "signed_out"})
	})
}
