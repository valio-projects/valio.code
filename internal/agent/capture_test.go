package agent

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repo(t *testing.T) string {
	t.Helper()
	if _, e := exec.LookPath("git"); e != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	git(t, root, "init", "-q")
	return root
}
func git(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if _, e := cmd.Output(); e != nil {
		t.Fatal("test Git command failed", e)
	}
}
func write(t *testing.T, root, path, content string) {
	t.Helper()
	p := filepath.Join(root, path)
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, []byte(content), 0600); e != nil {
		t.Fatal(e)
	}
}
func capture(t *testing.T, root string) Snapshot {
	t.Helper()
	s, e := Capture(context.Background(), Options{Root: root})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func filePaths(s Snapshot) map[string]bool {
	m := map[string]bool{}
	for _, f := range s.Files {
		m[f.Path] = true
	}
	return m
}

func TestGitIgnoreTrackedAndUntracked(t *testing.T) {
	r := repo(t)
	write(t, r, "tracked.log", "tracked\n")
	git(t, r, "add", "tracked.log")
	write(t, r, ".gitignore", "*.log\nignored/\n")
	write(t, r, "untracked.log", "ignored\n")
	write(t, r, "ignored/x.go", "package ignored\n")
	write(t, r, "main.go", "package main\n")
	s := capture(t, r)
	paths := filePaths(s)
	if !paths["tracked.log"] || !paths["main.go"] || paths["untracked.log"] || paths["ignored/x.go"] {
		t.Fatalf("unexpected paths: %v", paths)
	}
}
func TestSecretsNeverSerializedOrSpooled(t *testing.T) {
	r := repo(t)
	secrets := []string{"superSensitiveValue87654", "multilineSensitive_893745", "this_is_private_928475", "template_secret_348764", "json_password_789432", "yaml_secret_777333", "ghp_abcdefghijklmnopqrstuvwx"}
	write(t, r, ".env", "API_KEY="+secrets[0]+"\nMULTILINE=\"start\n"+secrets[1]+"\nFAKE_KEY=inside_value\nend\"\nPUBLIC=value\n")
	write(t, r, "id_rsa", "-----BEGIN PRIVATE KEY-----\n"+secrets[2])
	write(t, r, ".env.example", "TOKEN="+secrets[3])
	write(t, r, "settings.json", `{"password":"`+secrets[4]+`","nested":{"url":"hidden"}}`)
	write(t, r, "settings.yaml", "payload: |\n  "+secrets[5]+"\n  fake: scalar-value")
	write(t, r, "main.go", "package main\n// "+secrets[6])
	write(t, r, "safe.go", "package safe\n")
	spool := filepath.Join(r, ".valio", "spool")
	s, e := Capture(context.Background(), Options{Root: r, SpoolDir: spool})
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(s)
	stored, e := os.ReadFile(filepath.Join(spool, s.ID+".json"))
	if e != nil {
		t.Fatal(e)
	}
	for _, secret := range secrets {
		if strings.Contains(string(b), secret) || strings.Contains(string(stored), secret) {
			t.Fatal("secret found in sanitized payload")
		}
	}
	if strings.Contains(string(b), "FAKE_KEY") {
		t.Fatal("multiline value interpreted as key")
	}
	if !strings.Contains(string(b), "API_KEY") {
		t.Fatal("missing configuration key metadata")
	}
	if len(s.Files) != 1 || s.Files[0].Path != "safe.go" {
		t.Fatalf("unexpected source files: %v", filePaths(s))
	}
	if e = Save(spool, s); e != nil {
		t.Fatal(e)
	}
	recaptured, e := Capture(context.Background(), Options{Root: r, SpoolDir: spool})
	if e != nil || recaptured.ID != s.ID {
		t.Fatal("spool changed snapshot identity")
	}
	ids, e := Pending(spool)
	if e != nil || len(ids) != 1 {
		t.Fatal("idempotent spool failed")
	}
	loaded, e := Load(spool, s.ID)
	if e != nil || loaded.ID != s.ID {
		t.Fatal("resume failed")
	}
}
func TestDeterministicBinaryOversizeAndValioIgnore(t *testing.T) {
	r := repo(t)
	write(t, r, ".valioignore", "/dist/\n**/*.tmp\n!nested/keep.tmp\n")
	write(t, r, "dist/source.go", "excluded")
	write(t, r, "nested/drop.tmp", "excluded")
	write(t, r, "nested/keep.tmp", "keep")
	write(t, r, "binary.dat", "binary\x00data")
	write(t, r, "huge.txt", strings.Repeat("x", int(DefaultMaxFileBytes)+1))
	a := capture(t, r)
	b := capture(t, r)
	if a.ID != b.ID {
		t.Fatal("capture is nondeterministic")
	}
	paths := filePaths(a)
	if paths["dist/source.go"] || paths["nested/drop.tmp"] || !paths["nested/keep.tmp"] || paths["binary.dat"] || paths["huge.txt"] {
		t.Fatalf("unexpected paths: %v", paths)
	}
	codes := map[string]bool{}
	for _, d := range a.Diagnostics {
		codes[d.Code] = true
	}
	if !codes["binary-excluded"] || !codes["oversize-excluded"] {
		t.Fatal("missing diagnostics")
	}
}
func TestSymlinkEscapeExcluded(t *testing.T) {
	r := repo(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if e := os.WriteFile(outside, []byte("never_read_external_secret"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(outside, filepath.Join(r, "link.txt")); e != nil {
		t.Skip("symlinks unavailable")
	}
	s := capture(t, r)
	b, _ := json.Marshal(s)
	if strings.Contains(string(b), "never_read_external_secret") || filePaths(s)["link.txt"] {
		t.Fatal("external symlink content escaped")
	}
}
func TestSpoolIntegrityAndErrors(t *testing.T) {
	r := repo(t)
	write(t, r, "main.go", "package main")
	s := capture(t, r)
	dir := t.TempDir()
	if e := Save(dir, s); e != nil {
		t.Fatal(e)
	}
	write(t, dir, s.ID+".json", `{"content":"secret-in-corrupt-payload"}`)
	_, e := Load(dir, s.ID)
	if e == nil || strings.Contains(e.Error(), "secret-in-corrupt-payload") {
		t.Fatal("corrupt payload must fail without content disclosure")
	}
	s.Files[0].Content = "tampered"
	if Save(dir, s) == nil {
		t.Fatal("mutated immutable snapshot accepted")
	}
}
