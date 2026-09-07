package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

func fixture(scope typeinfo.TypeScope, id string) typeinfo.TypeDescriptor {
	ev := typeinfo.EvidenceRef{ID: "declaration-" + id, Producer: "test-compiler"}
	return typeinfo.TypeDescriptor{ID: id, Scope: scope, Name: typeinfo.KnownFact("Widget", scope, ev), FullyQualifiedName: typeinfo.KnownFact("Example.Widget", scope, ev), Language: typeinfo.KnownFact("csharp", scope, ev), Kind: typeinfo.KnownFact(typeinfo.TypeClass, scope, ev), Layout: typeinfo.UnresolvedLayout(scope, "compiler not available")}
}
func scope() typeinfo.TypeScope {
	return typeinfo.TypeScope{WorkspaceID: "w", ProjectID: "p", BuildProfileID: "debug", VersionID: "v1"}
}
func TestNameLookupPreservesScopeCollisionsAndOverloads(t *testing.T) {
	a := fixture(scope(), "a")
	a.Methods = []typeinfo.MethodDescriptor{{ID: "method-int", Name: a.Name, Signature: typeinfo.KnownFact("M(int)", a.Scope, a.Name.Evidence...)}, {ID: "method-string", Name: a.Name, Signature: typeinfo.KnownFact("M(string)", a.Scope, a.Name.Evidence...)}}
	b := fixture(scope(), "b")
	b.Scope.ProjectID = "other"
	b = fixture(b.Scope, "b")
	c := fixture(scope(), "c")
	c.Scope.BuildProfileID = "release"
	c = fixture(c.Scope, "c")
	d := fixture(scope(), "d")
	d.Scope.VersionID = "v2"
	d = fixture(d.Scope, "d")
	catalog, err := NewCatalog([]typeinfo.TypeDescriptor{d, c, b, a})
	if err != nil {
		t.Fatal(err)
	}
	result, err := catalog.ByName(QueryScope{WorkspaceID: "w"}, "Widget")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MatchAmbiguous || len(result.Candidates) != 4 || result.Reference.Resolution != typeinfo.ReferenceCandidate {
		t.Fatalf("lost ambiguous candidates: %+v", result)
	}
	result, err = catalog.ByName(QueryScope{WorkspaceID: "w", ProjectID: "p", BuildProfileID: "debug", VersionID: "v1"}, "Example.Widget")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MatchExact || result.Candidates[0].Counts.Methods != 2 {
		t.Fatal("overloads merged or exact lookup failed")
	}
	*result.Candidates[0].Type.Name.Value = "mutated"
	again, _ := catalog.ByName(QueryScope{WorkspaceID: "w", ProjectID: "p", BuildProfileID: "debug", VersionID: "v1"}, "Widget")
	if again.Status != MatchExact {
		t.Fatal("caller mutated catalog")
	}
	missing, _ := catalog.ByName(QueryScope{WorkspaceID: "unrelated"}, "Widget")
	if missing.Status != MatchNotFound {
		t.Fatal("workspace isolation failed")
	}
}
func TestAbsentLayoutNeverBecomesZero(t *testing.T) {
	d := fixture(scope(), "a")
	d.Layout = typeinfo.TypeLayout{}
	catalog, err := NewCatalog([]typeinfo.TypeDescriptor{d})
	if err != nil {
		t.Fatal(err)
	}
	result, _ := catalog.ByName(QueryScope{WorkspaceID: "w"}, "Widget")
	layout := result.Candidates[0].Type.Layout
	if layout.SizeBytes.Value != nil || layout.AlignmentBytes.Value != nil || layout.SizeBytes.EffectiveStatus() != typeinfo.FactUnresolved {
		t.Fatal("missing layout treated as zero")
	}
	b, _ := json.Marshal(layout)
	if !strings.Contains(string(b), `"sizeBytes":{"status":"unresolved","value":null`) {
		t.Fatalf("unexpected JSON %s", b)
	}
}
func TestAttributesPreservedAtAllScopes(t *testing.T) {
	d := fixture(scope(), "a")
	ev := d.Name.Evidence[0]
	attribute := func(id string) typeinfo.AttributeUse {
		return typeinfo.AttributeUse{ID: id, Name: typeinfo.KnownFact("Example.AuditAttribute", d.Scope, ev), AttributeClass: typeinfo.SymbolReference{Resolution: typeinfo.ReferenceExact, Exact: &typeinfo.SymbolIdentity{ID: "audit-class", Scope: d.Scope}}, PositionalArguments: []typeinfo.AttributeValue{{Kind: typeinfo.AttributeLiteral, Type: typeinfo.TypeReference{Name: typeinfo.KnownFact("System.String", d.Scope, ev)}, Literal: typeinfo.KnownFact("hello", d.Scope, ev)}, {Kind: typeinfo.AttributeExpression, Expression: typeinfo.KnownFact("nameof(Widget)", d.Scope, ev)}, {Kind: typeinfo.AttributeRedacted, RedactionReason: "sensitive configuration"}}, NamedArguments: []typeinfo.NamedAttributeArgument{{Name: typeinfo.KnownFact("Enabled", d.Scope, ev), Value: typeinfo.AttributeValue{Kind: typeinfo.AttributeLiteral, Type: typeinfo.TypeReference{Name: typeinfo.KnownFact("System.Boolean", d.Scope, ev)}, Literal: typeinfo.KnownFact("true", d.Scope, ev)}}}}
	}
	d.Attributes = []typeinfo.AttributeUse{attribute("type-attr")}
	d.Fields = []typeinfo.FieldDescriptor{{ID: "f", Attributes: []typeinfo.AttributeUse{attribute("field-attr")}}}
	d.Properties = []typeinfo.PropertyDescriptor{{ID: "p", Attributes: []typeinfo.AttributeUse{attribute("property-attr")}}}
	d.Constructors = []typeinfo.MethodDescriptor{{ID: "ctor", Attributes: []typeinfo.AttributeUse{attribute("ctor-attr")}}}
	d.Methods = []typeinfo.MethodDescriptor{{ID: "m", Attributes: []typeinfo.AttributeUse{attribute("method-attr")}, Parameters: []typeinfo.ParameterDescriptor{{ID: "arg", Attributes: []typeinfo.AttributeUse{attribute("parameter-attr")}}}, Returns: []typeinfo.ReturnDescriptor{{ID: "return", Attributes: []typeinfo.AttributeUse{attribute("return-attr")}, Type: typeinfo.TypeReference{Attributes: []typeinfo.AttributeUse{attribute("return-type-attr")}}}}}}
	catalog, err := NewCatalog([]typeinfo.TypeDescriptor{d})
	if err != nil {
		t.Fatal(err)
	}
	result, _ := catalog.ByName(QueryScope{WorkspaceID: "w"}, "Widget")
	candidate := result.Candidates[0]
	if candidate.Counts.Attributes != 8 {
		t.Fatalf("attribute scopes lost: %+v", candidate.Counts)
	}
	got := candidate.Type.Methods[0].Returns[0].Type.Attributes[0]
	if got.AttributeClass.Exact.ID != "audit-class" || len(got.PositionalArguments) != 3 || *got.NamedArguments[0].Value.Literal.Value != "true" {
		t.Fatal("attribute arguments or class target lost")
	}
	d.Attributes[0].PositionalArguments[2].Literal = typeinfo.KnownFact("secret", d.Scope, ev)
	if _, err := NewCatalog([]typeinfo.TypeDescriptor{d}); err == nil {
		t.Fatal("redacted attribute retained payload")
	}
}
func TestEnumOccurrencesRemainPerMemberAndPhysicalLayoutNeedsEvidence(t *testing.T) {
	d := fixture(scope(), "enum")
	ev := d.Name.Evidence[0]
	occurrence := func(id string) typeinfo.OccurrenceLink {
		return typeinfo.OccurrenceLink{ID: id, Range: domain.SourceRange{RepositoryID: "r", Path: "enums.cs", Encoding: domain.EncodingUTF8, Start: domain.Position{Line: 1}, End: domain.Position{Line: 1, Character: 1}}, Role: typeinfo.KnownFact("use", d.Scope, ev)}
	}
	d.Enum = &typeinfo.EnumDescriptor{UnderlyingType: typeinfo.TypeReference{Name: typeinfo.KnownFact("System.Byte", d.Scope, ev)}, Members: []typeinfo.EnumMember{{ID: "red", Name: typeinfo.KnownFact("Red", d.Scope, ev), Value: typeinfo.ConstantValue{Value: typeinfo.KnownFact("0", d.Scope, ev)}, Occurrences: []typeinfo.OccurrenceLink{occurrence("red-use-1"), occurrence("red-use-2")}}, {ID: "blue", Name: typeinfo.KnownFact("Blue", d.Scope, ev), Value: typeinfo.ConstantValue{Value: typeinfo.UnresolvedFact[string](d.Scope, "compiler unavailable")}, Occurrences: []typeinfo.OccurrenceLink{occurrence("blue-use-1")}}}}
	catalog, err := NewCatalog([]typeinfo.TypeDescriptor{d})
	if err != nil {
		t.Fatal(err)
	}
	result, _ := catalog.ByName(QueryScope{WorkspaceID: "w"}, "Widget")
	got := result.Candidates[0]
	if got.Counts.EnumMembers != 2 || got.Counts.Occurrences != 3 || got.Type.Enum.Members[1].Occurrences[0].ID != "blue-use-1" {
		t.Fatal("enum occurrences combined")
	}
	if got.Type.Enum.Members[1].Value.Value.Value != nil {
		t.Fatal("unresolved enum value treated as zero")
	}
	d.Layout.SizeBytes = typeinfo.KnownFact(uint64(1), d.Scope, ev)
	if d.Validate() == nil {
		t.Fatal("accepted size without target/compiler evidence")
	}
}
