package structure

import "testing"

func hash(t *testing.T, s string) Result {
	t.Helper()
	r, e := Fingerprint(s)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestShapeIsNotEquivalence(t *testing.T) {
	a := hash(t, "package p; func F() int { return 1 }")
	b := hash(t, "package p; func F() int { return 999 }")
	if a.Hash != b.Hash {
		t.Fatal("literal abstraction missing")
	}
	if a.Equivalence != "unknown" || a.Evidence != "syntax-shape" {
		t.Fatal("structural match claimed equivalence")
	}
}
func TestPreservesOperatorsControlAndBindingNames(t *testing.T) {
	base := hash(t, "package p; func F(x int) int { return x+1 }")
	for _, s := range []string{"package p; func F(x int) int { return x-1 }", "package p; func F(y int) int { return y+1 }", "package p; func F(x int) int { if x>0 { return x+1 }; return x }"} {
		if hash(t, s).Hash == base.Hash {
			t.Fatalf("lost semantic structure %s", s)
		}
	}
}
func TestFormattingAndCommentsIgnored(t *testing.T) {
	a := hash(t, "package p; func F() int { return 1 }")
	b := hash(t, "package p\n// comment\nfunc F() int {\n return 2 // foo\n}\n")
	if a.Hash != b.Hash {
		t.Fatal("format affected fingerprint")
	}
}
func TestInvalidSyntaxRejected(t *testing.T) {
	if _, e := Fingerprint("package p; func ("); e == nil {
		t.Fatal("invalid source fingerprinted")
	}
}
