package syntax

import (
	"encoding/json"
	"testing"
)

func TestReportRejectsOutOfBoundsOrInventedSemanticFacts(t *testing.T) {
	for _, fragment := range []string{`{"start":0,"end":99}`, `{"start":0.5,"end":1}`, `{"start":0,"end":1,"resolution":"exact"}`} {
		raw := json.RawMessage(`{"schema":1,"path":"a.cs","language":"csharp","validSyntax":true,"symbols":[` + fragment + `],"references":[],"imports":[],"diagnostics":[],"capabilities":{"syntax":true,"semanticResolution":false}}`)
		if validateReport(raw, "a.cs", "csharp", 10) == nil {
			t.Fatal("invalid source assertion accepted")
		}
	}
	raw := json.RawMessage(`{"schema":1,"path":"a.cs","language":"csharp","validSyntax":true,"symbols":[{"start":0,"end":1}],"references":[],"imports":[],"diagnostics":[],"capabilities":{"syntax":true,"semanticResolution":false}}`)
	if err := validateReport(raw, "a.cs", "csharp", 10); err != nil {
		t.Fatal(err)
	}
}

func TestUnavailableParserCannotPublishSyntaxCapability(t *testing.T) {
	raw := json.RawMessage(`{"schema":1,"path":"a.cs","language":"csharp","validSyntax":false,"symbols":[],"references":[],"imports":[],"diagnostics":[],"capabilities":{"syntax":false,"semanticResolution":false}}`)
	if validateReport(raw, "a.cs", "csharp", 10) == nil {
		t.Fatal("unavailable parser accepted as analysis")
	}
}
