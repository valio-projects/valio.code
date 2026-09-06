package telemetry

import (
	"context"
	"go.uber.org/fx"
)

// Module provides lifecycle-managed telemetry configuration and tracing.
var Module = fx.Module("telemetry", fx.Provide(ConfigFromEnvironment, provide), fx.Invoke(func(*Provider) {}))

func provide(lifecycle fx.Lifecycle, c Config) (*Provider, error) {
	p, err := NewProvider(context.Background(), c)
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: p.Shutdown})
	return p, nil
}
