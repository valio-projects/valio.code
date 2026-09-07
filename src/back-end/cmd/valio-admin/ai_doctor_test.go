package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	profileconfig "github.com/valio-projects/valio.code/internal/configuration/ai"
)

func TestAIDoctorRejectsBadPathAndUnknownProfileBeforeProbing(t *testing.T) {
	called := false
	probe := func(context.Context) []profileconfig.LocalProbe {
		called = true
		return nil
	}
	if err := runAIDoctorWith([]string{"--config", filepath.Join(t.TempDir(), "missing"), "--profile", "known"}, probe, &bytes.Buffer{}); err == nil || called {
		t.Fatal("bad profile path reached networking")
	}
	path := writeDoctorProfile(t)
	if err := runAIDoctorWith([]string{"--config", path, "--profile", "unknown"}, probe, &bytes.Buffer{}); err == nil || called {
		t.Fatal("unknown profile reached networking")
	}
}

func TestAIDoctorListsProfilesWithoutTokenEnvironment(t *testing.T) {
	path := writeDoctorProfile(t)
	var output bytes.Buffer
	if err := runAIDoctorWith([]string{"--config", path, "--list"}, func(context.Context) []profileconfig.LocalProbe {
		t.Fatal("list mode reached networking")
		return nil
	}, &output); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("profile=known")) || bytes.Contains(output.Bytes(), []byte("DO_NOT_PRINT")) {
		t.Fatalf("unsanitized list output: %q", output.String())
	}
}

func TestAIDoctorUsesConfiguredEmbeddingTimeout(t *testing.T) {
	document := profileconfig.Document{Profiles: []profileconfig.Profile{{ID: "slow", Timeout: "45s"}}}
	timeout, err := profileTimeout(document, "slow")
	if err != nil || timeout.Seconds() != 45 {
		t.Fatalf("profile timeout = %v, %v", timeout, err)
	}
	document.Profiles[0].Timeout = "121s"
	if _, err := profileTimeout(document, "slow"); err == nil {
		t.Fatal("timeout beyond the profile maximum was accepted")
	}
}

func TestAIDoctorSeparatesDiscoveryFromEmbeddingDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"embeddings":[[1,2]]}`))
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "profiles.json")
	body := fmt.Sprintf(`{"schemaVersion":1,"profiles":[{"id":"known","providerKind":"ollama","url":%q,"allowInsecureHTTP":true,"model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"3s","kind":"code","schemaVersion":1}]}`, server.URL)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := runAIDoctorWith([]string{"--config", path, "--profile", "known"}, func(ctx context.Context) []profileconfig.LocalProbe {
		<-ctx.Done()
		return nil
	}, &output)
	if err != nil || !strings.Contains(output.String(), "profile=known status=available") {
		t.Fatalf("doctor after exhausted discovery context: output=%q err=%v", output.String(), err)
	}
}

func writeDoctorProfile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "profiles.json")
	body := `{"schemaVersion":1,"profiles":[{"id":"known","providerKind":"ollama","url":"http://127.0.0.1:11434","allowInsecureHTTP":true,"model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"1s","tokenEnv":"DO_NOT_PRINT","kind":"code","schemaVersion":1}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
