package telemetry

import "os"

// Config controls OpenTelemetry service identity and optional OTLP export.
type Config struct {
	// ServiceName identifies traces; empty environment input defaults to valio-code.
	ServiceName string
	// Endpoint is an optional HTTP(S) OTLP base endpoint.
	Endpoint string
}

// ConfigFromEnvironment reads supported OTEL settings without exposing their values.
func ConfigFromEnvironment() Config {
	name := os.Getenv("OTEL_SERVICE_NAME")
	if name == "" {
		name = "valio-code"
	}
	return Config{ServiceName: name, Endpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")}
}
