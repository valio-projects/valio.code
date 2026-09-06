package telemetry

import "os"

type Config struct {
	ServiceName string
	Endpoint    string
}

func ConfigFromEnvironment() Config {
	name := os.Getenv("OTEL_SERVICE_NAME")
	if name == "" {
		name = "valio-code"
	}
	return Config{ServiceName: name, Endpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")}
}
