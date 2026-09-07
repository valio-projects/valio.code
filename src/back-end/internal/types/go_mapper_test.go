package types

import (
	"testing"

	"github.com/valio-projects/valio.code/internal/analysis"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

func TestGoMapperPreservesRealMembersAndUnknownLayout(t *testing.T) {
	source := "package sample\ntype Embedded struct{}\ntype Widget[T ~int] struct { Embedded; Name string `json:\"name\"`; Values [2][3]int }\nfunc (w *Widget[T]) Method(value T, rest ...string) (string, error) { return w.Name, nil }\n"
	report := analysis.AnalyzeWithOptions("widget.go", source, analysis.Options{CheckTypes: true})
	descriptors, err := FromGoReport(scope(), "repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog(descriptors)
	if err != nil {
		t.Fatal(err)
	}
	result, err := catalog.ByName(QueryScope{WorkspaceID: "w"}, "Widget")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MatchExact {
		t.Fatalf("unexpected result %+v", result)
	}
	typ := result.Candidates[0].Type
	if len(typ.Fields) != 3 || len(typ.Methods) != 1 || len(typ.GenericParameters) != 1 {
		t.Fatalf("missing Go metadata: %+v", result.Candidates[0].Counts)
	}
	if *typ.Fields[1].Tag.Value != `json:"name"` || len(typ.Fields[1].Attributes) != 0 {
		t.Fatal("Go struct tag converted into an attribute")
	}
	method := typ.Methods[0]
	if len(method.Parameters) != 2 || len(method.Returns) != 2 || method.Receiver == nil || *method.Parameters[1].Modifiers.Value == nil {
		t.Fatal("signature lost arguments, returns, receiver or variadic modifier")
	}
	if typ.Layout.SizeBytes.Value != nil || typ.Layout.AlignmentBytes.Value != nil {
		t.Fatal("mapper guessed physical layout")
	}
	array := typ.Fields[2].Type.Array
	if array == nil || *array.Rank.Value != 2 || *array.Dimensions[0].Length.Value != 2 || *array.Dimensions[1].Length.Value != 3 || *array.Dimensions[0].LowerBound.Value != 0 {
		t.Fatal("checked array dimensions lost")
	}
	if typ.Fields[0].Declaration.Range.Start.Line != 2 || typ.Fields[0].Declaration.Range.Encoding != domain.EncodingUTF8 {
		t.Fatal("source encoding/range lost")
	}
}
func TestGoMapperTypedConstantsAreNotEnumsAndKeepUsageIDs(t *testing.T) {
	source := "package sample\ntype Color int\nconst ( Red Color = iota; Blue )\nvar colors = []Color{Red, Blue, Red}\n"
	report := analysis.AnalyzeWithOptions("colors.go", source, analysis.Options{CheckTypes: true})
	mapped, err := FromGoReport(scope(), "repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapped) != 1 {
		t.Fatalf("expected Color, got %d", len(mapped))
	}
	typ := mapped[0]
	if typ.Enum != nil || len(typ.Constants) != 2 {
		t.Fatal("typed Go constants incorrectly represented as enum")
	}
	red, blue := typ.Constants[0], typ.Constants[1]
	if *red.Value.Value.Value != "0" || *blue.Value.Value.Value != "1" || len(red.Occurrences) != 2 || len(blue.Occurrences) != 1 {
		t.Fatal("checker constants or per-member uses lost")
	}
	ids := map[string]bool{}
	for _, member := range typ.Constants {
		for _, use := range member.Occurrences {
			if ids[use.ID] || use.Symbol.Resolution != typeinfo.ReferenceExact || use.Symbol.Exact.ID != member.ID {
				t.Fatal("usage identity or exact target incorrect")
			}
			ids[use.ID] = true
		}
	}
	unchecked, err := FromGoReport(scope(), "repo", source, analysis.Analyze("colors.go", source))
	if err != nil {
		t.Fatal(err)
	}
	if unchecked[0].Constants[0].Value.Value.Value != nil || len(unchecked[0].Constants[0].Occurrences) != 0 {
		t.Fatal("AST-only report invented constant values/uses")
	}
	other, err := FromGoReport(scope(), "other-repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	if other[0].ID == typ.ID {
		t.Fatal("same paths in different repositories collided")
	}
}
func TestGoMapperRejectsMismatchedSourceRanges(t *testing.T) {
	source := "package sample\ntype Widget struct{}\n"
	report := analysis.Analyze("widget.go", source)
	if _, err := FromGoReport(scope(), "repo", "short", report); err == nil {
		t.Fatal("accepted source too short for report")
	}
}
