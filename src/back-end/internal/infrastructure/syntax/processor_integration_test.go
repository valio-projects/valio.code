package syntax

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestConfiguredHelperLanguages(t *testing.T) {
	path := os.Getenv("VALIO_TEST_SYNTAX_HELPER")
	if path == "" {
		t.Skip("requires installed pinned syntax helper")
	}
	processor, err := New("node", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ language, path, content, name string }{
		{"c", "sample.c", "struct User { int id; }; int read_id(struct User u) { return u.id; }", "User"},
		{"cpp", "sample.cpp", "class User { public: int id; virtual int get() { return id; } };", "User"},
		{"csharp", "sample.cs", "public class User { public int ID {get;set;} public virtual int Get(int n) {return n;} }", "User"},
		{"javascript", "sample.js", "export class User { get(n) { return n; } }", "User"},
		{"typescript", "sample.ts", "export class User { id: number; get(n: number): number {return n;} }", "User"},
		{"java", "User.java", "public class User { public int id; public int get(int n) { return n; } }", "User"},
	} {
		t.Run(test.language, func(t *testing.T) {
			raw, err := processor.Analyze(context.Background(), test.path, test.language, test.content)
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				Valid   bool `json:"validSyntax"`
				Symbols []struct {
					Name string `json:"name"`
				} `json:"symbols"`
			}
			if json.Unmarshal(raw, &result) != nil || !result.Valid {
				t.Fatalf("invalid parse %s", raw)
			}
			found := false
			for _, symbol := range result.Symbols {
				if symbol.Name == test.name {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing expected %q declaration", test.name)
			}
		})
	}
}
