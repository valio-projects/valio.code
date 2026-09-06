package configgraph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnvMultilineAndTemplateValues(t *testing.T) {
	p := Extract(".env.example", "env", []byte("FIRST='secret\nNOT_A_KEY=stillsecret\nend'\nexport SECOND=anothersecret\n"))
	b, _ := json.Marshal(p)
	if len(p.Keys) != 2 || p.Keys[0].Name != "FIRST" || p.Keys[1].Name != "SECOND" || strings.Contains(string(b), "secret") {
		t.Fatalf("invalid safe projection: %s", b)
	}
}
func TestYAMLExcluded(t *testing.T) {
	p := Extract("config.yaml", "yaml", []byte("a: |\n  lookslike: secret\n"))
	if len(p.Keys) != 0 {
		t.Fatal("YAML projection must remain conservative")
	}
}
