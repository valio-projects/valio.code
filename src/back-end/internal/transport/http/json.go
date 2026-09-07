package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"io"
	"mime"
	"net/http"
)

func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, status int, code string) {
	write(w, status, map[string]any{"error": map[string]string{"code": code, "message": code}, "requestId": w.Header().Get("X-Request-ID")})
}
func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || media != "application/json" {
		failure(w, 415, "JSON_REQUIRED")
		return false
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e = d.Decode(value); e != nil {
		var limit *http.MaxBytesError
		if errors.As(e, &limit) {
			failure(w, 413, "SCOPE_TOO_LARGE")
		} else {
			failure(w, 400, "INVALID_JSON")
		}
		return false
	}
	if e = d.Decode(new(any)); e != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(e, &limit) {
			failure(w, 413, "SCOPE_TOO_LARGE")
		} else {
			failure(w, 400, "INVALID_JSON")
		}
		return false
	}
	return true
}
func respond(w http.ResponseWriter, value any, e error) {
	if e == nil {
		write(w, 200, value)
		return
	}
	switch {
	case errors.Is(e, fault.ErrUnavailable):
		failure(w, 503, "CAPABILITY_UNAVAILABLE")
	case errors.Is(e, fault.ErrNotFound):
		failure(w, 404, "NOT_FOUND")
	case errors.Is(e, fault.ErrForbidden):
		failure(w, 403, "WORKSPACE_FORBIDDEN")
	case errors.Is(e, fault.ErrInvalid):
		failure(w, 400, "INVALID_REQUEST")
	case errors.Is(e, fault.ErrScopeTooLarge):
		failure(w, 413, "SCOPE_TOO_LARGE")
	case errors.Is(e, fault.ErrConflict):
		failure(w, 409, "CONFLICT")
	case errors.Is(e, context.Canceled), errors.Is(e, context.DeadlineExceeded):
		failure(w, 408, "REQUEST_TIMEOUT")
	default:
		failure(w, 500, "INTERNAL_ERROR")
	}
}
