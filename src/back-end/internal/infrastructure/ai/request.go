package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/valio-projects/valio.code/internal/validation"
)

const maxRequestBytes = 1 << 20
const maxResponseBytes = 4 << 20

var errRedirect = errors.New("AI endpoint redirects are not allowed")

func request(ctx context.Context, config Config, path string, body any) ([]byte, error) {
	client, base, err := config.client()
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(body)
	if err != nil || len(payload) > maxRequestBytes {
		return nil, errors.New("AI request exceeds budget")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(payload))
	if err != nil {
		return nil, errors.New("AI request could not be created")
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Token != "" {
		if err := validation.Token(config.Token, 1); err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}
	response, err := client.Do(req)
	if err != nil {
		if errors.Is(err, errRedirect) {
			return nil, errRedirect
		}
		return nil, transportError(ctx)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, transportError(ctx)
	}
	if len(data) > maxResponseBytes {
		return nil, errors.New("AI response exceeds budget")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode > 299 {
		return nil, fmt.Errorf("AI endpoint status %d", response.StatusCode)
	}
	return data, nil
}

func transportError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("AI endpoint unavailable")
}
