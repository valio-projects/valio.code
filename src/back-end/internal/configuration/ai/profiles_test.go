package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsUnknownTrailingAndDuplicateProfiles(t *testing.T) {
	for _, body := range []string{`{"schemaVersion":1,"profiles":[{"id":"a","providerKind":"ollama","url":"http://127.0.0.1:11434","model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"1s","kind":"code","schemaVersion":1,"unknown":true}]}`, `{"schemaVersion":1,"profiles":[]} {}`, `{"schemaVersion":1,"profiles":[{"id":"a","providerKind":"ollama","url":"http://127.0.0.1:11434","model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"1s","kind":"code","schemaVersion":1},{"id":"a","providerKind":"ollama","url":"http://127.0.0.1:11434","model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"1s","kind":"code","schemaVersion":1}]}`} {
		p := filepath.Join(t.TempDir(), "p.json")
		if e := os.WriteFile(p, []byte(body), 0600); e != nil {
			t.Fatal(e)
		}
		if _, e := Load(p); e == nil {
			t.Fatal("accepted invalid JSON profile")
		}
	}
}

func TestCheckedInExampleLoads(t *testing.T) {
	document, err := Load("../../../../../deploy/ai-profiles.example.json")
	if err != nil || len(document.Profiles) != 9 {
		t.Fatalf("example did not load: %v", err)
	}
	quantized := 0
	for _, profile := range document.Profiles {
		if profile.ProviderKind == "ollama" {
			quantized++
			if !profile.AllowInsecureHTTP || profile.Quantization == "" || profile.Dimension < 1 {
				t.Fatalf("invalid checked-in quantized profile %q", profile.ID)
			}
		}
	}
	if quantized != 6 {
		t.Fatalf("quantized profiles = %d, want 6", quantized)
	}
	if _, _, err = document.Resolve("qwen3-8b-lmstudio-docker"); err != nil {
		t.Fatalf("host Docker profile with explicit plaintext permission did not resolve: %v", err)
	}
	if _, _, err = document.Resolve("qwen3-8b-lmstudio-local"); err != nil {
		t.Fatalf("native LM Studio profile did not resolve: %v", err)
	}
}

func TestLoadRejectsDockerPlaintextWithoutExplicitPermission(t *testing.T) {
	body := `{"schemaVersion":1,"profiles":[{"id":"docker","providerKind":"openai-compatible","url":"http://host.docker.internal:1234/v1","model":"m","revision":"r","quantization":"q","dimension":2,"timeout":"1s","kind":"code","schemaVersion":1}]}`
	path := filepath.Join(t.TempDir(), "p.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("accepted Docker plaintext endpoint without explicit permission")
	}
}
