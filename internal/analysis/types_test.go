package analysis

import (
	"strings"
	"testing"
)

func typeNamed(t *testing.T, r Report, name string) TypeInfo {
	t.Helper()
	for _, typ := range r.Types {
		if typ.Name == name {
			return typ
		}
	}
	t.Fatalf("type %q missing", name)
	return TypeInfo{}
}
func TestRichTypesFieldsMethodsAndSignatures(t *testing.T) {
	src := "package rich\ntype Base struct{}\ntype Box[T any] struct {\n *Base\n Value T `json:\"value,omitempty\"`\n hidden *[2][3]int\n A, B string\n}\nfunc (b *Box[T]) Read(value *T, extras ...string) (out T, err error) { return }\nfunc (b Box[T]) private() {}\nfunc New[T any](value T) *Box[T] { return &Box[T]{Value:value} }\ntype Alias = *Box[int]\ntype Reader interface { Read(p []byte) (n int, err error) }\n"
	r := (Parser{}).Analyze("rich.go", src)
	if !r.ValidSyntax {
		t.Fatal(r.Diagnostics)
	}
	b := typeNamed(t, r, "Box")
	if b.Kind != "struct" || !b.Exported || len(b.TypeParameters) != 1 || len(b.Fields) != 5 || len(b.Methods) != 2 {
		t.Fatalf("bad Box: %+v", b)
	}
	if !b.Fields[0].Embedded || b.Fields[0].Name != "Base" || src[b.Fields[0].NameRange.Start:b.Fields[0].NameRange.End] != "Base" {
		t.Fatalf("bad embedded field: %+v", b.Fields[0])
	}
	if b.Fields[1].Tag != `json:"value,omitempty"` || b.Fields[1].Type.Text != "T" {
		t.Fatalf("tag/type lost: %+v", b.Fields[1])
	}
	hidden := b.Fields[2]
	if hidden.Exported || hidden.Visibility != "package" || hidden.Type.Kind != "pointer" || len(hidden.Type.Element.Dimensions) != 2 {
		t.Fatalf("bad private array: %+v", hidden)
	}
	if hidden.Type.Element.Dimensions[0].Length != nil {
		t.Fatal("syntax guessed array dimension")
	}
	m := b.Methods[0]
	if m.Receiver.Type.Text != "*Box[T]" || len(m.Parameters) != 2 || len(m.Returns) != 2 || m.Returns[0].Name != "out" || !m.Parameters[1].Variadic || !strings.Contains(m.Signature, "Read(value *T, extras ...string)") {
		t.Fatalf("bad signature: %+v", m)
	}
	if b.Methods[1].Exported || b.Layout.Status != "unknown" || b.Layout.SizeBytes != nil || b.Layout.AlignmentBytes != nil {
		t.Fatal("visibility/layout incorrectly inferred")
	}
	if len(r.Functions) != 1 || r.Functions[0].Name != "New" || r.Functions[0].Returns[0].Type.Text != "*Box[T]" {
		t.Fatalf("bad functions: %+v", r.Functions)
	}
	alias := typeNamed(t, r, "Alias")
	if !alias.Alias || alias.Kind != "alias" || alias.UnderlyingType.Text != "*Box[int]" {
		t.Fatalf("bad alias %+v", alias)
	}
	reader := typeNamed(t, r, "Reader")
	if len(reader.Methods) != 1 || reader.Methods[0].Range.Start == 0 {
		t.Fatalf("interface method range %+v", reader)
	}
}

func TestSameNamedFieldsAndLocalTypesStayDistinct(t *testing.T) {
	src := `package p
type A struct { Same int }
type B struct { Same string }
func f() { type A struct { Local bool }; _ = A{} }
`
	r := Analyze("same.go", src)
	if len(r.Types) != 2 {
		t.Fatalf("local type leaked into package catalog: %v", r.Types)
	}
	a, b := typeNamed(t, r, "A"), typeNamed(t, r, "B")
	if a.Fields[0].ID == b.Fields[0].ID || a.Fields[0].Type.Text == b.Fields[0].Type.Text {
		t.Fatal("same-name fields were conflated")
	}
}

func TestActualTypeCheckerConstantsAndShadowing(t *testing.T) {
	src := `package typed
type State int
const (
 Ready State = iota + 1
 Done
)
const Other State = 7
type Matrix [2+1][4]int
func read() State { return Ready }
func local() int { Ready := 9; return Ready }
`
	r := AnalyzeWithOptions("typed.go", src, Options{CheckTypes: true, PackagePath: "example/typed"})
	if r.TypeCheckStatus != "complete" || !r.Capabilities.TypeChecked {
		t.Fatalf("checker did not succeed: %+v", r.TypeDiagnostics)
	}
	state := typeNamed(t, r, "State")
	if len(state.Constants) != 3 || state.ResolvedUnderlyingType != "int" {
		t.Fatalf("bad typed constants: %+v", state)
	}
	ready, done, other := state.Constants[0], state.Constants[1], state.Constants[2]
	if ready.Value != "1" || done.Value != "2" || other.Value != "7" || !done.InheritedExpression || ready.GroupID != done.GroupID || other.GroupID == ready.GroupID {
		t.Fatalf("constant group/value lost: %+v", state.Constants)
	}
	if ready.Resolution != "go-types" || len(ready.References) != 1 || src[ready.References[0].Range.Start:ready.References[0].Range.End] != "Ready" {
		t.Fatalf("constant references conflated shadowing: %+v", ready)
	}
	if len(r.ConstantGroups) != 2 || r.ConstantGroups[0].Kind != "go-const-group" {
		t.Fatalf("Go const group mislabeled: %+v", r.ConstantGroups)
	}
	m := typeNamed(t, r, "Matrix")
	dims := m.UnderlyingType.Dimensions
	if len(dims) != 2 || dims[0].Length == nil || *dims[0].Length != 3 || *dims[1].Length != 4 {
		t.Fatalf("checker dimensions missing: %+v", dims)
	}
	if m.Layout.Status != "unknown" || m.Layout.SizeBytes != nil {
		t.Fatal("type checking invented target ABI")
	}
}

func TestSyntaxConstantReferencesStayUnresolved(t *testing.T) {
	src := `package p; type Flag int; const Yes Flag = 1; var x = Yes`
	r := Analyze("p.go", src)
	c := typeNamed(t, r, "Flag").Constants[0]
	if c.Value != "" || len(c.References) != 0 || c.Resolution != "unresolved" || r.TypeCheckStatus != "not-run" {
		t.Fatalf("syntax asserted constant semantics: %+v", c)
	}
}

func TestPartialTypeCheckDoesNotClaimComplete(t *testing.T) {
	r := AnalyzeWithOptions("missing.go", "package p; type T Missing; var x = unknown", Options{CheckTypes: true})
	if !r.ValidSyntax || r.TypeCheckStatus != "partial" || r.Capabilities.TypeChecked || len(r.TypeDiagnostics) == 0 {
		t.Fatalf("missing checker diagnostics: %+v", r)
	}
	for _, o := range r.Occurrences {
		if o.Name == "unknown" && o.Resolution != "unresolved" {
			t.Fatal("missing symbol was resolved")
		}
	}
}

func TestInvalidArrayLengthsRemainUnresolved(t *testing.T) {
	for _, dimension := range []string{`"bad"`, `-1`, `true`} {
		r := AnalyzeWithOptions("array.go", "package p; type T ["+dimension+"]int", Options{CheckTypes: true})
		typ := typeNamed(t, r, "T")
		if r.TypeCheckStatus != "partial" || typ.UnderlyingType.Dimensions[0].Length != nil {
			t.Fatalf("invalid array dimension treated as known: %+v", typ)
		}
	}
}
