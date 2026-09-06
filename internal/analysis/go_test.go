package analysis

import "testing"

func TestDefinitionsAndUnresolvedUses(t *testing.T) {
	source := `package demo
import "fmt"
type Box[T any] struct { Value T }
func (b Box[T]) Read(n int) int {
 x := n
 x, y := x + 1, 2
 fmt.Println(b.Value, y)
 return x
}
`
	r := Analyze("demo.go", source)
	if !r.ValidSyntax {
		t.Fatalf("%+v", r.Diagnostics)
	}
	if r.Capabilities.TypeChecked || r.Capabilities.ResolvedCalls || r.Capabilities.ResolvedReferences {
		t.Fatal("overclaimed capabilities")
	}
	defs := map[string]int{}
	for _, s := range r.Symbols {
		if source[s.Range.Start:s.Range.End] != s.Name {
			t.Fatalf("bad range %+v", s)
		}
		defs[s.Name]++
		if s.Evidence != "go-ast-syntax" {
			t.Fatal("missing evidence")
		}
	}
	if defs["Read"] != 1 || defs["x"] != 1 || defs["y"] != 1 || defs["Value"] != 1 {
		t.Fatalf("incorrect definitions %v", defs)
	}
	uses := 0
	for _, o := range r.Occurrences {
		if o.Role == "use" {
			uses++
			if o.Resolution != "unresolved" || o.SymbolID != "" {
				t.Fatalf("invented reference: %+v", o)
			}
		}
	}
	if uses == 0 {
		t.Fatal("missing occurrences")
	}
	if len(r.Imports) != 1 || r.Imports[0].Path != "fmt" {
		t.Fatalf("bad imports: %+v", r.Imports)
	}
}
func TestSyntaxErrorsProduceDiagnostics(t *testing.T) {
	r := Analyze("broken.go", "package p\nfunc Broken( {\n")
	if r.ValidSyntax || len(r.Diagnostics) == 0 {
		t.Fatalf("missing syntax failure %+v", r)
	}
	if r.Capabilities.TypeChecked {
		t.Fatal("invalid file typechecked")
	}
}
func TestUnicodeByteRangesAndCRLF(t *testing.T) {
	src := "package p\r\nfunc 世界(参数 int) int { return 参数 }\r\n"
	r := Analyze("unicode.go", src)
	if !r.ValidSyntax {
		t.Fatal(r.Diagnostics)
	}
	for _, s := range r.Symbols {
		if src[s.Range.Start:s.Range.End] != s.Name {
			t.Fatalf("not byte offsets: %+v", s)
		}
	}
}
