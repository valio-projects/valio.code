package agent

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func reidentify(s *Snapshot) {
	s.ID = ""
	b, _ := json.Marshal(s)
	sum := sha256.Sum256(b)
	s.ID = hex.EncodeToString(sum[:])
}
func TestUploadSanitizedPayloadAndAuthorization(t *testing.T) {
	r := repo(t)
	write(t, r, ".env", "PASSWORD=do_not_upload_this_value\n")
	write(t, r, "main.go", "package main\n")
	s := capture(t, r)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ingestion" || r.Method != "POST" {
			t.Error("wrong ingestion request")
		}
		if r.Header.Get("Authorization") != "Bearer test-credential" {
			t.Error("missing bearer")
		}
		if r.Header.Get("Idempotency-Key") != s.ID {
			t.Error("missing idempotency key")
		}
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "do_not_upload_this_value") {
			t.Error("raw configuration value uploaded")
		}
		var payload struct {
			WorkspaceID  string   `json:"workspaceId"`
			RepositoryID string   `json:"repositoryId"`
			Snapshot     Snapshot `json:"snapshot"`
		}
		if json.Unmarshal(b, &payload) != nil || payload.WorkspaceID != "w" || payload.RepositoryID != "r" || ValidateSnapshot(payload.Snapshot) != nil {
			t.Error("invalid upload contract")
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(UploadResult{SnapshotID: s.ID, JobID: "job-1", Status: "accepted"})
	}))
	defer server.Close()
	result, e := Upload(context.Background(), UploadOptions{Endpoint: server.URL, Token: "test-credential", WorkspaceID: "w", RepositoryID: "r"}, s)
	if e != nil || result.SnapshotID != s.ID {
		t.Fatal("upload failed", e)
	}
}
func TestUploadRefusesNonLocalHTTPAndSanitizesErrors(t *testing.T) {
	for _, url := range []string{"http://example.com", "http://127.0.0.2", "http://localhost.example.com", "https://user:credential@example.com", "https://example.com?token=secret"} {
		if _, e := ingestionEndpoint(url); e == nil {
			t.Fatal("unsafe endpoint accepted")
		}
	}
	for _, url := range []string{"http://localhost:1234", "http://127.0.0.1:1234", "http://[::1]:1234", "https://example.com"} {
		if _, e := ingestionEndpoint(url); e != nil {
			t.Fatal("safe endpoint rejected")
		}
	}
	r := repo(t)
	write(t, r, "safe.go", "package safe")
	s := capture(t, r)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		io.WriteString(w, "do_not_echo_server_secret")
	}))
	defer server.Close()
	_, e := Upload(context.Background(), UploadOptions{server.URL, "hidden_token", "w", "r"}, s)
	if e == nil || strings.Contains(e.Error(), "do_not_echo") || strings.Contains(e.Error(), "hidden_token") {
		t.Fatal("unsafe upload error")
	}
}
func TestTLSRedirectCannotForwardToken(t *testing.T) {
	r := repo(t)
	write(t, r, "safe.go", "package safe")
	s := capture(t, r)
	redirected := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected = true }))
	defer target.Close()
	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/steal", http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	roots := x509.NewCertPool()
	roots.AddCert(source.Certificate())
	transport.TLSClientConfig = source.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	transport.TLSClientConfig.RootCAs = roots
	defer transport.CloseIdleConnections()
	_, e := (UploadClient{Options: UploadOptions{source.URL, "never-forward-me", "w", "r"}, Transport: transport}).Upload(context.Background(), s)
	if e == nil || redirected {
		t.Fatal("authenticated redirect followed")
	}
}
func TestRehashedUnsafeSnapshotsRejected(t *testing.T) {
	r := repo(t)
	write(t, r, "main.go", "package main\n")
	original := capture(t, r)
	for _, mutate := range []func(*Snapshot){
		func(s *Snapshot) { s.Files[0].Path = ".env" },
		func(s *Snapshot) {
			s.Files[0].Content = "password = privatevalue"
			s.Files[0].Size = len(s.Files[0].Content)
			sum := sha256.Sum256([]byte(s.Files[0].Content))
			s.Files[0].Hash = hex.EncodeToString(sum[:])
		},
		func(s *Snapshot) { s.Files[0].Path = "../outside.go" },
		func(s *Snapshot) {
			s.Diagnostics = append(s.Diagnostics, Diagnostic{Path: "main.go", Code: "unreadable", Message: "secret raw line"})
		},
	} {
		data, _ := json.Marshal(original)
		var changed Snapshot
		json.Unmarshal(data, &changed)
		mutate(&changed)
		reidentify(&changed)
		if ValidateSnapshot(changed) == nil || Save(t.TempDir(), changed) == nil {
			t.Fatal("rehashing bypassed safety policy")
		}
	}
}
