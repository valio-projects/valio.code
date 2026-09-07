package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/valio-projects/valio.code/internal/analysis"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// GoMapper attaches project/version scope and provenance to actual parser and
// checker output. It does not infer compiler layout or resolve lexical names.
type GoMapper struct {
	scope      typeinfo.TypeScope
	repository domain.RepositoryID
	source     string
	lineStarts []int
}

func NewGoMapper(scope typeinfo.TypeScope, repository domain.RepositoryID, source string) (*GoMapper, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	if repository == "" {
		return nil, fmt.Errorf("Go mapper requires repository identity")
	}
	starts := []int{0}
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return &GoMapper{scope: scope, repository: repository, source: source, lineStarts: starts}, nil
}
func FromGoReport(scope typeinfo.TypeScope, repository domain.RepositoryID, source string, report analysis.Report) ([]typeinfo.TypeDescriptor, error) {
	m, err := NewGoMapper(scope, repository, source)
	if err != nil {
		return nil, err
	}
	return m.Map(report)
}
func (m *GoMapper) id(local string) string {
	h := sha256.Sum256([]byte(string(m.repository) + "\x00" + local))
	return hex.EncodeToString(h[:16])
}
func (m *GoMapper) exact(local string) typeinfo.SymbolReference {
	return typeinfo.SymbolReference{Resolution: typeinfo.ReferenceExact, Exact: &typeinfo.SymbolIdentity{ID: m.id(local), Scope: m.scope}, Candidates: []typeinfo.SymbolIdentity{}}
}
func (m *GoMapper) sourceRange(path string, r analysis.Range) (domain.SourceRange, error) {
	if r.Start < 0 || r.End < r.Start || r.End > len(m.source) {
		return domain.SourceRange{}, fmt.Errorf("analysis range [%d,%d) is outside supplied source", r.Start, r.End)
	}
	position := func(offset int) domain.Position {
		line := sort.Search(len(m.lineStarts), func(i int) bool { return m.lineStarts[i] > offset }) - 1
		return domain.Position{Line: line, Character: offset - m.lineStarts[line]}
	}
	return domain.SourceRange{RepositoryID: m.repository, Path: path, Encoding: domain.EncodingUTF8, Start: position(r.Start), End: position(r.End)}, nil
}
func (m *GoMapper) evidence(path, id, producer string, r analysis.Range) (typeinfo.EvidenceRef, error) {
	sr, err := m.sourceRange(path, r)
	if err != nil {
		return typeinfo.EvidenceRef{}, err
	}
	return typeinfo.EvidenceRef{ID: m.id(id + ":" + producer), Producer: producer, Range: &sr}, nil
}
func (m *GoMapper) occurrence(path, id, role, symbol string, r analysis.Range, exact bool) (*typeinfo.OccurrenceLink, error) {
	sr, err := m.sourceRange(path, r)
	if err != nil {
		return nil, err
	}
	ev, err := m.evidence(path, id, "go-ast-syntax", r)
	if err != nil {
		return nil, err
	}
	ref := typeinfo.SymbolReference{Resolution: typeinfo.ReferenceUnresolved, Reason: "semantic resolution unavailable", Candidates: []typeinfo.SymbolIdentity{}}
	if exact && symbol != "" {
		ref = m.exact(symbol)
	}
	return &typeinfo.OccurrenceLink{ID: m.id(id), Range: sr, Role: typeinfo.KnownFact(role, m.scope, ev), Symbol: ref}, nil
}
func (m *GoMapper) Map(report analysis.Report) ([]typeinfo.TypeDescriptor, error) {
	if err := domain.ValidateRelativePath(report.Path, false); err != nil {
		return nil, err
	}
	out := []typeinfo.TypeDescriptor{}
	for _, t := range report.Types {
		if t.ID == "" {
			return nil, fmt.Errorf("Go type %q has no identity", t.Name)
		}
		ev, err := m.evidence(report.Path, t.ID, "go-ast-syntax", t.NameRange)
		if err != nil {
			return nil, err
		}
		decl, err := m.occurrence(report.Path, t.ID+":declaration", "definition", t.ID, t.NameRange, true)
		if err != nil {
			return nil, err
		}
		d := typeinfo.TypeDescriptor{ID: m.id(t.ID), Scope: m.scope, Name: typeinfo.KnownFact(t.Name, m.scope, ev), FullyQualifiedName: typeinfo.KnownFact(t.QualifiedName, m.scope, ev), Language: typeinfo.KnownFact("go", m.scope, ev), Kind: typeinfo.KnownFact(typeinfo.TypeKind(t.Kind), m.scope, ev), Visibility: typeinfo.KnownFact(goVisibility(t.Visibility), m.scope, ev), Modifiers: typeinfo.KnownFact([]typeinfo.Modifier{}, m.scope, ev), Symbol: m.exact(t.ID), Declaration: decl, Fields: []typeinfo.FieldDescriptor{}, Properties: []typeinfo.PropertyDescriptor{}, Methods: []typeinfo.MethodDescriptor{}, Constructors: []typeinfo.MethodDescriptor{}, GenericParameters: []typeinfo.GenericParameter{}, Attributes: []typeinfo.AttributeUse{}, Constants: []typeinfo.EnumMember{}, EmbeddedTypes: []typeinfo.TypeReference{}, Layout: typeinfo.UnresolvedLayout(m.scope, "no target/compiler/ABI evidence; Go source cannot establish physical layout")}
		underlying := m.typeReference(t.UnderlyingType, ev)
		d.UnderlyingType = &underlying
		d.Array = underlying.Array
		for _, embedded := range t.EmbeddedTypes {
			d.EmbeddedTypes = append(d.EmbeddedTypes, m.typeReference(embedded, ev))
		}
		for _, f := range t.Fields {
			field, err := m.field(report.Path, f)
			if err != nil {
				return nil, err
			}
			d.Fields = append(d.Fields, field)
		}
		for _, fn := range t.Methods {
			method, err := m.method(report.Path, fn)
			if err != nil {
				return nil, err
			}
			d.Methods = append(d.Methods, method)
		}
		for i, p := range t.TypeParameters {
			d.GenericParameters = append(d.GenericParameters, m.generic(t.ID, i, p, ev))
		}
		for _, c := range t.Constants {
			member, err := m.constant(report.Path, c)
			if err != nil {
				return nil, err
			}
			d.Constants = append(d.Constants, member)
		}
		if err := d.Validate(); err != nil {
			return nil, fmt.Errorf("mapped Go type %q: %w", t.Name, err)
		}
		out = append(out, d)
	}
	return out, nil
}
func goVisibility(value string) typeinfo.Visibility {
	if value == "exported" {
		return typeinfo.VisibilityPublic
	}
	if value == "package" {
		return typeinfo.VisibilityPackage
	}
	return typeinfo.Visibility(value)
}
func (m *GoMapper) typeReference(t analysis.TypeExpression, ev typeinfo.EvidenceRef) typeinfo.TypeReference {
	ref := typeinfo.TypeReference{Name: typeinfo.KnownFact(t.Text, m.scope, ev), Symbol: typeinfo.SymbolReference{Resolution: typeinfo.ReferenceUnresolved, Reason: "written type expression; no scoped semantic target supplied", Candidates: []typeinfo.SymbolIdentity{}}}
	if t.Kind == "array" {
		element := t.Element
		for element != nil && element.Kind == "array" {
			element = element.Element
		}
		array := &typeinfo.ArrayShape{Rank: typeinfo.KnownFact(len(t.Dimensions), m.scope, ev), Dimensions: []typeinfo.ArrayDimension{}}
		if element != nil {
			array.ElementType = m.typeReference(*element, ev)
		}
		for _, dim := range t.Dimensions {
			length := typeinfo.UnresolvedFact[uint64](m.scope, "array length expression has not been resolved: "+dim.Expression)
			if dim.Length != nil && *dim.Length >= 0 && dim.Resolution == "go-types" {
				checker := ev
				checker.Producer = "go-types"
				checker.ID += "-checked"
				length = typeinfo.KnownFact(uint64(*dim.Length), m.scope, checker)
			}
			array.Dimensions = append(array.Dimensions, typeinfo.ArrayDimension{Length: length, LowerBound: typeinfo.KnownFact(int64(0), m.scope, ev)})
		}
		ref.Array = array
	}
	return ref
}
func (m *GoMapper) generic(parent string, index int, p analysis.ParameterInfo, ev typeinfo.EvidenceRef) typeinfo.GenericParameter {
	return typeinfo.GenericParameter{ID: m.id(fmt.Sprintf("%s:generic:%d", parent, index)), Name: typeinfo.KnownFact(p.Name, m.scope, ev), Constraints: []typeinfo.TypeReference{m.typeReference(p.Type, ev)}, Attributes: []typeinfo.AttributeUse{}}
}
func (m *GoMapper) parameter(parent string, index int, p analysis.ParameterInfo, ev typeinfo.EvidenceRef) typeinfo.ParameterDescriptor {
	modifiers := []typeinfo.Modifier{}
	if p.Variadic {
		modifiers = append(modifiers, typeinfo.ModifierVariadic)
	}
	return typeinfo.ParameterDescriptor{ID: m.id(fmt.Sprintf("%s:parameter:%d", parent, index)), Name: typeinfo.KnownFact(p.Name, m.scope, ev), Position: typeinfo.KnownFact(index, m.scope, ev), Type: m.typeReference(p.Type, ev), Modifiers: typeinfo.KnownFact(modifiers, m.scope, ev), DefaultValue: typeinfo.UnsupportedFact[string](m.scope, "Go parameters do not have default arguments"), Attributes: []typeinfo.AttributeUse{}}
}

// Provenance stores source text only in syntax facts. Redaction of application
// secrets must occur before a source blob/report enters this mapper.
func nonemptyName(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
