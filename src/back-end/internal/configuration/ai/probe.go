package ai

import (
	"context"
	"net/http"
	"time"
)

// ProbeStatus reports whether a read-only local provider endpoint responded.
type ProbeStatus string

const (
	// ProbeAvailable means the endpoint returned a 2xx status to its model-list request.
	ProbeAvailable ProbeStatus = "available"
	// ProbeAbsent means the endpoint could not be reached or did not return a 2xx status.
	ProbeAbsent ProbeStatus = "absent"
)

// LocalProbe is a bounded readiness observation that contains no response body,
// model contents, token, or configuration value.
type LocalProbe struct {
	// Provider identifies the local API dialect that was probed.
	Provider string `json:"provider"`
	// Endpoint is the fixed loopback model-list endpoint.
	Endpoint string `json:"endpoint"`
	// Status records whether the endpoint responded successfully.
	Status ProbeStatus `json:"status"`
}

// ProbeLocal sends read-only model-list requests to the conventional local
// Ollama and LM Studio endpoints. It never starts a service or downloads a model.
func ProbeLocal(ctx context.Context) []LocalProbe {
	client := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return []LocalProbe{
		probeEndpoint(ctx, client, "ollama", "http://localhost:11434/api/tags"),
		probeEndpoint(ctx, client, "lm-studio", "http://localhost:1234/v1/models"),
	}
}

func probeEndpoint(ctx context.Context, client *http.Client, provider, endpoint string) LocalProbe {
	result := LocalProbe{Provider: provider, Endpoint: endpoint, Status: ProbeAbsent}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return result
	}
	response, err := client.Do(req)
	if err != nil {
		return result
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		result.Status = ProbeAvailable
	}
	return result
}
