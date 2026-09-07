package syntaxfacts

import (
	"encoding/json"
	"testing"
)

func validReport() Report {
	return Report{Schema: 1, Language: "typescript", Path: "user.ts", ValidSyntax: true, Capabilities: Capabilities{Syntax: true}, Symbols: []Declaration{{ID: "class", Name: "User", Kind: "class", Start: 0, End: 8}, {ID: "field", Name: "id", Kind: "field", ParentID: "class", Start: 3, End: 5}}}
}

func TestDecodeChecksPinnedScopeAndUTF8Ranges(t *testing.T) {
	source := "abcédef" // Eight UTF-8 bytes; byte 4 is inside a rune.
	report := validReport()
	raw, _ := json.Marshal(report)
	if _, err := Decode(raw, "user.ts", "typescript", source); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(raw, "other.ts", "typescript", source); err == nil {
		t.Fatal("different source accepted")
	}
	report.Symbols[1].Start = 4
	if err := report.Validate(source); err == nil {
		t.Fatal("range inside UTF-8 rune accepted")
	}
}

func TestValidateRejectsForgedDeclarationRelationships(t *testing.T) {
	cases := []func(*Report){
		func(r *Report) { r.Symbols[1].ID = "class" },
		func(r *Report) { r.Symbols[0].ParentID = "field" },
		func(r *Report) { r.Symbols[1].ParentID = "missing" },
		func(r *Report) { r.Symbols[1].Parameters = []Parameter{{Start: 0, End: 2}} },
		func(r *Report) { r.References = []Reference{{Start: 0, End: 1, Resolution: "exact"}} },
		func(r *Report) { r.Capabilities.SemanticResolution = true },
	}
	for index, mutate := range cases {
		r := validReport()
		mutate(&r)
		if r.Validate("abcdefgh") == nil {
			t.Fatalf("forged relationship case %d accepted", index)
		}
	}
}
