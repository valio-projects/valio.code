package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/validation"
	"io"
	"net/http"
	"strings"
	"time"
)

// UploadOptions identifies an authenticated snapshot-ingestion destination.
type UploadOptions struct {
	// Endpoint is HTTPS or loopback HTTP ingestion URL.
	Endpoint string
	// Token is sent only as the endpoint bearer credential.
	Token string
	// WorkspaceID scopes the destination workspace.
	WorkspaceID string
	// RepositoryID identifies the destination repository.
	RepositoryID string
}

// UploadResult is the server receipt for an accepted snapshot.
type UploadResult struct {
	// SnapshotID identifies the accepted snapshot.
	SnapshotID string `json:"snapshotId"`
	// ViewID identifies an optional resulting analysis view.
	ViewID string `json:"viewId,omitempty"`
	// JobID identifies optional asynchronous processing.
	JobID string `json:"jobId,omitempty"`
	// Status is the server-reported receipt status.
	Status string `json:"status"`
}

// UploadClient uploads snapshots with an optional custom HTTP transport.
type UploadClient struct {
	// Options supplies endpoint, credential, and destination identities.
	Options UploadOptions
	// Transport optionally replaces the default transport.
	Transport http.RoundTripper
}

// Upload sends only validated sanitized snapshots. It does not remove durable
// spool data: ambiguous network outcomes can be retried using the same identity.
func Upload(ctx context.Context, opts UploadOptions, s Snapshot) (UploadResult, error) {
	return (UploadClient{Options: opts}).Upload(ctx, s)
}

// Upload validates s, sends it, and returns a sanitized server receipt.
func (u UploadClient) Upload(ctx context.Context, s Snapshot) (UploadResult, error) {
	opts := u.Options
	if err := opts.Validate(); err != nil {
		return UploadResult{}, err
	}
	if err := ValidateSnapshot(s); err != nil {
		return UploadResult{}, err
	}
	endpoint, err := ingestionEndpoint(opts.Endpoint)
	if err != nil {
		return UploadResult{}, err
	}
	if err := validation.Token(opts.Token, 1); err != nil {
		return UploadResult{}, err
	}
	if opts.WorkspaceID == "" || opts.RepositoryID == "" {
		return UploadResult{}, errors.New("workspace and repository IDs are required")
	}
	payload, err := json.Marshal(struct {
		WorkspaceID  string   `json:"workspaceId"`
		RepositoryID string   `json:"repositoryId"`
		Snapshot     Snapshot `json:"snapshot"`
	}{opts.WorkspaceID, opts.RepositoryID, s})
	if err != nil {
		return UploadResult{}, errors.New("cannot encode upload payload")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return UploadResult{}, errors.New("cannot create upload request")
	}
	request.Header.Set("Authorization", "Bearer "+opts.Token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", s.ID)
	client := &http.Client{Transport: u.Transport, Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return UploadResult{}, ctx.Err()
		}
		return UploadResult{}, errors.New("snapshot upload failed; sanitized snapshot retained in spool")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusAccepted {
		return UploadResult{}, errors.New("server did not accept snapshot; sanitized snapshot retained in spool")
	}
	// Never include server bodies or transport errors in diagnostics: either can
	// echo credentials, source values, or an authenticated redirect destination.
	var result UploadResult
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if readErr != nil || len(body) > 64*1024 || json.Unmarshal(body, &result) != nil || result.SnapshotID == "" || result.Status == "" {
		return UploadResult{}, errors.New("server returned an invalid upload receipt; snapshot retained in spool")
	}
	return result, nil
}
func ingestionEndpoint(raw string) (string, error) {
	invalid := func() (string, error) {
		return "", errors.New("server must use HTTPS (HTTP allowed only for localhost, 127.0.0.1, or ::1)")
	}
	u, e := validation.Endpoint(raw, true)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return invalid()
	}
	if u.Scheme != "https" {
		if u.Scheme != "http" {
			return invalid()
		}
		switch strings.ToLower(u.Hostname()) {
		case "localhost", "127.0.0.1", "::1":
		default:
			return invalid()
		}
	}
	if !strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/api/v1/ingestion") {
		u.Path = strings.TrimRight(u.Path, "/") + "/api/v1/ingestion"
	} else {
		u.Path = strings.TrimRight(u.Path, "/")
	}
	u.RawPath = ""
	return u.String(), nil
}
