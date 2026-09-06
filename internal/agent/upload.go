package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type UploadOptions struct {
	Endpoint     string
	Token        string
	WorkspaceID  string
	RepositoryID string
}
type UploadResult struct {
	SnapshotID string `json:"snapshotId"`
	ViewID     string `json:"viewId,omitempty"`
	JobID      string `json:"jobId,omitempty"`
	Status     string `json:"status"`
}

type UploadClient struct {
	Options   UploadOptions
	Transport http.RoundTripper
}

// Upload sends only validated sanitized snapshots. It does not remove durable
// spool data: ambiguous network outcomes can be retried using the same identity.
func Upload(ctx context.Context, opts UploadOptions, s Snapshot) (UploadResult, error) {
	return (UploadClient{Options: opts}).Upload(ctx, s)
}

func (u UploadClient) Upload(ctx context.Context, s Snapshot) (UploadResult, error) {
	opts := u.Options
	if err := ValidateSnapshot(s); err != nil {
		return UploadResult{}, err
	}
	endpoint, err := ingestionEndpoint(opts.Endpoint)
	if err != nil {
		return UploadResult{}, err
	}
	if strings.TrimSpace(opts.Token) == "" || strings.ContainsAny(opts.Token, "\r\n") {
		return UploadResult{}, errors.New("VALIO_API_TOKEN is required and must be a single-line token")
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
		return UploadResult{}, errors.New("snapshot upload failed; sanitized snapshot retained in spool")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusAccepted {
		return UploadResult{}, errors.New("server did not accept snapshot; sanitized snapshot retained in spool")
	}
	// Never include server bodies or transport errors in diagnostics: either can
	// echo credentials, source values, or an authenticated redirect destination.
	var result UploadResult
	if json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&result) != nil || result.SnapshotID == "" || result.Status == "" {
		return UploadResult{}, errors.New("server returned an invalid upload receipt; snapshot retained in spool")
	}
	return result, nil
}
func ingestionEndpoint(raw string) (string, error) {
	invalid := func() (string, error) {
		return "", errors.New("server must use HTTPS (HTTP allowed only for localhost, 127.0.0.1, or ::1)")
	}
	u, e := url.Parse(raw)
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
