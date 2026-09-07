package types

import (
	"strings"
	"testing"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

func TestSyntaxMapperMapsCSharpMembersAttributesAndNestedTypes(t *testing.T) {
	source := "[Serializable]\npublic class Outer { [FieldAttr] protected internal int Shared; public string Run([ArgAttr] ref string input) { return input; } public class Nested { private int Count; } }\npublic enum State : byte { [EnumAttr] Ready = 1 << 2, Waiting }\n"
	outerStart := strings.Index(source, "public class Outer")
	nestedStart := strings.Index(source, "public class Nested")
	enumStart := strings.Index(source, "public enum State")
	report := syntaxReport("csharp", "state.cs", []syntaxfacts.Declaration{
		declaration("outer", "Outer", "class", "", outerStart, nestedEnd(source, outerStart), "public", []string{"public"}, []string{"[Serializable]"}),
		withType(declaration("shared", "Shared", "field", "outer", strings.Index(source, "[FieldAttr]"), strings.Index(source, "; public string"), "protected internal", []string{"protected", "internal"}, []string{"[FieldAttr]"}), "int", nil),
		withType(declaration("run", "Run", "method", "outer", strings.Index(source, "public string Run"), strings.Index(source, " public class Nested"), "public", []string{"public"}, nil), "string", []syntaxfacts.Parameter{parameter("input", "string", strings.Index(source, "[ArgAttr]"), strings.Index(source, ") { return"), []string{"ref"}, []string{"[ArgAttr]"})}),
		declaration("nested", "Nested", "class", "outer", nestedStart, strings.Index(source, " } }\npublic enum")+2, "public", []string{"public"}, nil),
		withType(declaration("count", "Count", "field", "nested", strings.Index(source, "private int Count"), strings.Index(source, "; } }"), "private", []string{"private"}, nil), "int", nil),
		withUnderlying(declaration("state", "State", "enum", "", enumStart, strings.LastIndex(source, " }\n")+2, "public", []string{"public"}, nil), "byte"),
		withEnumValue(declaration("ready", "Ready", "enum_member", "state", strings.Index(source, "[EnumAttr]"), strings.Index(source, ", Waiting"), "unknown", nil, []string{"[EnumAttr]"}), "1 << 2"),
		declaration("waiting", "Waiting", "enum_member", "state", strings.Index(source, "Waiting"), strings.LastIndex(source, " }\n")+1, "unknown", nil, nil),
	})

	mapped, err := FromSyntaxReport(scope(), domain.RepositoryID("repo"), source, report)
	if err != nil {
		t.Fatal(err)
	}
	outer := syntaxType(t, mapped, "Outer")
	if outer.FullyQualifiedName.EffectiveStatus() != typeinfo.FactUnresolved || outer.Layout.SizeBytes.Value != nil {
		t.Fatal("mapper invented semantic identity or layout")
	}
	if outer.Visibility.Value == nil || *outer.Visibility.Value != typeinfo.VisibilityPublic || len(outer.Attributes) != 1 || outer.Attributes[0].RawSyntax == nil || outer.Attributes[0].RawSyntax.Value == nil || *outer.Attributes[0].RawSyntax.Value != "[Serializable]" || outer.Attributes[0].Name.EffectiveStatus() != typeinfo.FactUnresolved || len(outer.Attributes[0].PositionalArguments) != 0 || len(outer.Attributes[0].NamedArguments) != 0 {
		t.Fatalf("type syntax facts lost: %+v", outer)
	}
	if len(outer.Fields) != 1 || *outer.Fields[0].Visibility.Value != typeinfo.VisibilityProtectedInternal || *outer.Fields[0].Type.Name.Value != "int" || len(outer.Fields[0].Attributes) != 1 {
		t.Fatalf("C# field facts lost: %+v", outer.Fields)
	}
	method := outer.Methods[0]
	if len(method.Parameters) != 1 || *method.Parameters[0].Type.Name.Value != "string" || (*method.Parameters[0].Modifiers.Value)[0] != typeinfo.Modifier("ref") || len(method.Parameters[0].Attributes) != 1 || len(method.Returns) != 1 || *method.Returns[0].Type.Name.Value != "string" {
		t.Fatalf("method signature facts lost: %+v", method)
	}
	nested := syntaxType(t, mapped, "Nested")
	if len(nested.Fields) != 1 || *nested.Fields[0].Name.Value != "Count" || nested.Symbol.Resolution != typeinfo.ReferenceUnresolved {
		t.Fatalf("nested type or its direct members were not retained as syntax facts: %+v", nested)
	}
	state := syntaxType(t, mapped, "State")
	if state.Enum == nil || state.Enum.UnderlyingType.Name.Value == nil || *state.Enum.UnderlyingType.Name.Value != "byte" || len(state.Enum.Members) != 2 || state.Enum.Members[0].Expression.Value == nil || *state.Enum.Members[0].Expression.Value != "1 << 2" || state.Enum.Members[0].Value.Value.Value != nil || state.Enum.Members[1].Value.Value.Value != nil {
		t.Fatalf("enum syntax facts lost or guessed: %+v", state.Enum)
	}
}

func TestSyntaxMapperPreservesCPlusPlusAndTypeScriptWrittenFacts(t *testing.T) {
	cpp := "class Counter { public: const int value; int Add(const int delta); };"
	cppReport := syntaxReport("cpp", "counter.cpp", []syntaxfacts.Declaration{
		declaration("counter", "Counter", "class", "", 0, len(cpp), "unknown", nil, nil),
		withType(declaration("value", "value", "field", "counter", strings.Index(cpp, "const int value"), strings.Index(cpp, "; int Add")+1, "public", []string{"const"}, nil), "int", nil),
		withType(declaration("add", "Add", "method", "counter", strings.Index(cpp, "int Add"), strings.Index(cpp, ");")+2, "public", nil, nil), "int", []syntaxfacts.Parameter{parameter("delta", "const int", strings.Index(cpp, "const int delta"), strings.Index(cpp, ")"), nil, nil)}),
	})
	mapped, err := FromSyntaxReport(scope(), "repo", cpp, cppReport)
	if err != nil {
		t.Fatal(err)
	}
	counter := syntaxType(t, mapped, "Counter")
	if *counter.Fields[0].Type.Name.Value != "int" || (*counter.Fields[0].Modifiers.Value)[0] != typeinfo.ModifierConst || *counter.Methods[0].Returns[0].Type.Name.Value != "int" {
		t.Fatalf("C++ written type/modifier facts lost: %+v", counter)
	}

	ts := "@sealed export class Service { public readonly endpoint: string; run(input: number): Promise<void> {} }"
	tsReport := syntaxReport("typescript", "service.ts", []syntaxfacts.Declaration{
		declaration("service", "Service", "class", "", 0, len(ts), "public", []string{"export"}, []string{"@sealed"}),
		withType(declaration("endpoint", "endpoint", "field", "service", strings.Index(ts, "public readonly endpoint"), strings.Index(ts, "; run")+1, "public", []string{"public", "readonly"}, nil), "string", nil),
		withType(declaration("run", "run", "method", "service", strings.Index(ts, "run(input"), strings.Index(ts, " {}")+3, "unknown", nil, nil), "Promise<void>", []syntaxfacts.Parameter{parameter("input", "number", strings.Index(ts, "input: number"), strings.Index(ts, "):"), nil, nil)}),
	})
	mapped, err = FromSyntaxReport(scope(), "repo", ts, tsReport)
	if err != nil {
		t.Fatal(err)
	}
	service := syntaxType(t, mapped, "Service")
	if len(service.Attributes) != 1 || *service.Fields[0].Type.Name.Value != "string" || (*service.Fields[0].Modifiers.Value)[1] != typeinfo.ModifierReadonly || *service.Methods[0].Parameters[0].Type.Name.Value != "number" {
		t.Fatalf("TypeScript syntax facts lost: %+v", service)
	}
}

func TestSyntaxMapperRejectsInvalidSourceBounds(t *testing.T) {
	source := "class C {}"
	report := syntaxReport("typescript", "bad.ts", []syntaxfacts.Declaration{declaration("c", "C", "class", "", 0, len(source)+1, "unknown", nil, nil)})
	if _, err := FromSyntaxReport(scope(), "repo", source, report); err == nil {
		t.Fatal("accepted report outside supplied source")
	}
}

func TestSyntaxMapperUsesUTF8ByteRanges(t *testing.T) {
	source := "// привет 😀\nclass C {}"
	start := strings.Index(source, "class C")
	report := syntaxReport("javascript", "unicode.js", []syntaxfacts.Declaration{declaration("c", "C", "class", "", start, len(source), "unknown", nil, nil)})
	mapped, err := FromSyntaxReport(scope(), "repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	declaration := mapped[0].Declaration
	if declaration.Range.Encoding != domain.EncodingUTF8 || declaration.Range.Start.Line != 1 || declaration.Range.Start.Character != 0 {
		t.Fatalf("UTF-8 source range was not preserved: %+v", declaration.Range)
	}
}

func TestSyntaxMapperLeavesMissingSemanticMetadataUnresolved(t *testing.T) {
	source := "interface I { Run(value); }"
	report := syntaxReport("javascript", "contract.js", []syntaxfacts.Declaration{
		declaration("interface", "I", "interface", "", 0, len(source), "unknown", nil, nil),
		declaration("run", "Run", "method", "interface", strings.Index(source, "Run"), strings.Index(source, ";")+1, "unknown", nil, nil),
	})
	report.Symbols[1].Parameters = []syntaxfacts.Parameter{parameter("value", "", strings.Index(source, "value"), strings.Index(source, ")"), nil, nil)}
	mapped, err := FromSyntaxReport(scope(), "repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := syntaxType(t, mapped, "I")
	method := descriptor.Methods[0]
	if descriptor.FullyQualifiedName.EffectiveStatus() != typeinfo.FactUnresolved || descriptor.Symbol.Resolution != typeinfo.ReferenceUnresolved || descriptor.Layout.Kind.EffectiveStatus() != typeinfo.FactUnresolved || method.Parameters[0].Type.Name.EffectiveStatus() != typeinfo.FactUnresolved || len(descriptor.GenericParameters) != 0 {
		t.Fatalf("mapper guessed unavailable semantic metadata: %+v", descriptor)
	}
}

func TestSyntaxMapperScopesHelperLocalIDsByPath(t *testing.T) {
	source := "[TypeAttr] class C { int M([ParameterAttr] int input) { return input; } }"
	mappedA, err := FromSyntaxReport(scope(), "repo", source, sameIDReport("first.cs", source))
	if err != nil {
		t.Fatal(err)
	}
	mappedB, err := FromSyntaxReport(scope(), "repo", source, sameIDReport("second.cs", source))
	if err != nil {
		t.Fatal(err)
	}
	first, second := mappedA[0], mappedB[0]
	firstMethod, secondMethod := first.Methods[0], second.Methods[0]
	if first.ID == second.ID || firstMethod.Parameters[0].ID == secondMethod.Parameters[0].ID || firstMethod.Returns[0].ID == secondMethod.Returns[0].ID || first.Attributes[0].ID == second.Attributes[0].ID || firstMethod.Parameters[0].Attributes[0].ID == secondMethod.Parameters[0].Attributes[0].ID {
		t.Fatalf("helper-local IDs collided across files: first=%+v second=%+v", first, second)
	}
}

func TestSyntaxMapperSkipsAnonymousTypeCatalogEntries(t *testing.T) {
	source := "class {}"
	report := syntaxReport("typescript", "anonymous.ts", []syntaxfacts.Declaration{declaration("s1", "", "class", "", 0, len(source), "unknown", nil, nil)})
	mapped, err := FromSyntaxReport(scope(), "repo", source, report)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapped) != 0 {
		t.Fatalf("anonymous syntax node became an invalid catalog descriptor: %+v", mapped)
	}
}

func syntaxReport(language, path string, symbols []syntaxfacts.Declaration) syntaxfacts.Report {
	return syntaxfacts.Report{Schema: 1, Language: language, Path: path, ValidSyntax: true, Symbols: symbols, References: []syntaxfacts.Reference{}, Imports: []syntaxfacts.Import{}, Diagnostics: []syntaxfacts.Diagnostic{}, Capabilities: syntaxfacts.Capabilities{Syntax: true, SemanticResolution: false}}
}

func sameIDReport(path, source string) syntaxfacts.Report {
	methodStart := strings.Index(source, "int M")
	methodEnd := strings.Index(source, " } }") + 2
	parameterStart := strings.Index(source, "[ParameterAttr]")
	parameterEnd := strings.Index(source, ") {")
	return syntaxReport("csharp", path, []syntaxfacts.Declaration{
		declaration("s1", "C", "class", "", 0, len(source), "unknown", nil, []string{"[TypeAttr]"}),
		withType(declaration("s2", "M", "method", "s1", methodStart, methodEnd, "unknown", nil, nil), "int", []syntaxfacts.Parameter{parameter("input", "int", parameterStart, parameterEnd, nil, []string{"[ParameterAttr]"})}),
	})
}

func declaration(id, name, kind, parent string, start, end int, visibility string, modifiers, attributes []string) syntaxfacts.Declaration {
	return syntaxfacts.Declaration{ID: id, Name: name, Kind: kind, ParentID: parent, Start: start, End: end, Visibility: visibility, Modifiers: modifiers, Attributes: attributes, Parameters: []syntaxfacts.Parameter{}}
}

func withType(value syntaxfacts.Declaration, typeName string, parameters []syntaxfacts.Parameter) syntaxfacts.Declaration {
	value.Type = &typeName
	value.Parameters = parameters
	return value
}

func withUnderlying(value syntaxfacts.Declaration, typeName string) syntaxfacts.Declaration {
	value.UnderlyingType = &typeName
	return value
}

func withEnumValue(value syntaxfacts.Declaration, expression string) syntaxfacts.Declaration {
	value.EnumValue = &expression
	return value
}

func parameter(name, typeName string, start, end int, modifiers, attributes []string) syntaxfacts.Parameter {
	return syntaxfacts.Parameter{Name: name, Type: &typeName, Start: start, End: end, Modifiers: modifiers, Attributes: attributes}
}

func syntaxType(t *testing.T, descriptors []typeinfo.TypeDescriptor, name string) typeinfo.TypeDescriptor {
	t.Helper()
	for _, descriptor := range descriptors {
		if descriptor.Name.Value != nil && *descriptor.Name.Value == name {
			return descriptor
		}
	}
	t.Fatalf("missing type %q", name)
	return typeinfo.TypeDescriptor{}
}

func nestedEnd(source string, start int) int {
	return strings.Index(source[start:], " } }\npublic enum") + start + 4
}
