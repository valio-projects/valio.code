package mcp

import "net/http"

// bearerTransport attaches credentials to a cloned outbound request.
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	request := r.Clone(r.Context())
	request.Header = r.Header.Clone()
	request.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(request)
}
