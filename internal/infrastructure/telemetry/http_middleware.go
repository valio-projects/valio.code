package telemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
)

// HTTPMiddleware uses official instrumentation with a sanitized request for
// telemetry. The application receives its original query and headers.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, original *http.Request) {
		instrumented := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, traced *http.Request) {
			next.ServeHTTP(w, original.WithContext(traced.Context()))
		}), "valio.http",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return "HTTP " + r.Method }))
		clean := original.Clone(original.Context())
		cleanURL := *original.URL
		cleanURL.RawQuery = ""
		cleanURL.ForceQuery = false
		cleanURL.User = nil
		clean.URL = &cleanURL
		clean.RequestURI = clean.URL.EscapedPath()
		// Header instrumentation only sees trace propagation and bounded protocol
		// metadata, never credentials, arbitrary headers, or user-agent contents.
		clean.Header = make(http.Header)
		for _, key := range []string{"Traceparent", "Tracestate", "Content-Type", "Content-Length"} {
			if value := original.Header.Get(key); value != "" {
				clean.Header.Set(key, value)
			}
		}
		instrumented.ServeHTTP(w, clean)
	})
}
