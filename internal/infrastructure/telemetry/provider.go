package telemetry

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Provider struct{ TracerProvider *sdktrace.TracerProvider }

func NewProvider(ctx context.Context, c Config) (*Provider, error) {
	options := []sdktrace.TracerProviderOption{sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", c.ServiceName))), sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample()))}
	if c.Endpoint != "" {
		u, err := url.Parse(c.Endpoint)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, errors.New("invalid OTLP endpoint")
		}
		exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(c.Endpoint), otlptracehttp.WithURLPath(strings.TrimRight(u.Path, "/")+"/v1/traces"), otlptracehttp.WithTimeout(5*time.Second))
		if err != nil {
			return nil, errors.New("cannot initialize OTLP exporter")
		}
		options = append(options, sdktrace.WithBatcher(exporter, sdktrace.WithMaxQueueSize(2048), sdktrace.WithBatchTimeout(time.Second)))
	}
	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)
	// Baggage is deliberately excluded: unreviewed attributes must not be
	// copied into jobs or exported to telemetry.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return &Provider{TracerProvider: provider}, nil
}
func (p *Provider) Shutdown(ctx context.Context) error { return p.TracerProvider.Shutdown(ctx) }
