package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndpointRejectsAmbiguousOrCredentialBearingAddresses(t *testing.T) {
	for _, raw := range []string{"", " https://example.com", "https://user:secret@example.com", "http://example.com", "https://example.com:0", "https://example.com:65536", "https://example.com:", "https://bad_host", "https://example.com?", "https://example.com#", "https://example.com/%2e%2e/path", "https://example.com/a%5Cb"} {
		if _, err := Endpoint(raw, true); err == nil {
			t.Errorf("accepted invalid URL %q", raw)
		} else if strings.Contains(err.Error(), "secret") {
			t.Fatal("credential echoed")
		}
	}
	for _, raw := range []string{"https://example.com/base", "http://localhost:8080", "http://127.0.0.1", "http://[::1]:8080"} {
		if _, err := Endpoint(raw, true); err != nil {
			t.Errorf("%s: %v", raw, err)
		}
	}
	if _, err := Endpoint("http://ollama:11434", false); err != nil {
		t.Fatal(err)
	}
}

func TestTokenRejectsHeaderInjectionAndWhitespace(t *testing.T) {
	for _, value := range []string{"", strings.Repeat("x", 31), strings.Repeat("x", 8193), strings.Repeat("x", 32) + "\x00", strings.Repeat("x", 32) + "\t", strings.Repeat("x", 32) + " ", strings.Repeat("x", 32) + "\r\nAuthorization: secret"} {
		if Token(value, 32) == nil {
			t.Fatal("invalid token accepted")
		}
	}
	if err := Token(strings.Repeat("x", 32), 32); err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryVerifiesExistenceAndType(t *testing.T) {
	root := t.TempDir()
	if actual, err := Directory(root); err != nil || !filepath.IsAbs(actual) {
		t.Fatal(actual, err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []string{"", filepath.Join(root, "missing"), file, "bad\x00path"} {
		if _, err := Directory(candidate); err == nil {
			t.Fatal("invalid directory accepted")
		}
	}
}
