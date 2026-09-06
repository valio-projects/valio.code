package types

import (
	"fmt"

	"github.com/valio-projects/valio.code/internal/analysis"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

func (m *GoMapper) field(path string, f analysis.FieldInfo) (typeinfo.FieldDescriptor, error) {
	ev, err := m.evidence(path, f.ID, "go-ast-syntax", f.Range)
	if err != nil {
		return typeinfo.FieldDescriptor{}, err
	}
	decl, err := m.occurrence(path, f.ID+":declaration", "definition", f.ID, f.NameRange, true)
	if err != nil {
		return typeinfo.FieldDescriptor{}, err
	}
	modifiers := []typeinfo.Modifier{}
	if f.Embedded {
		modifiers = append(modifiers, typeinfo.ModifierEmbedded)
	}
	return typeinfo.FieldDescriptor{ID: m.id(f.ID), Name: typeinfo.KnownFact(f.Name, m.scope, ev), Type: m.typeReference(f.Type, ev), Visibility: typeinfo.KnownFact(goVisibility(f.Visibility), m.scope, ev), Modifiers: typeinfo.KnownFact(modifiers, m.scope, ev), Tag: typeinfo.KnownFact(f.Tag, m.scope, ev), Attributes: []typeinfo.AttributeUse{}, Declaration: decl}, nil
}
func (m *GoMapper) method(path string, fn analysis.FunctionInfo) (typeinfo.MethodDescriptor, error) {
	ev, err := m.evidence(path, fn.ID, "go-ast-syntax", fn.Range)
	if err != nil {
		return typeinfo.MethodDescriptor{}, err
	}
	decl, err := m.occurrence(path, fn.ID+":declaration", "definition", fn.ID, fn.NameRange, true)
	if err != nil {
		return typeinfo.MethodDescriptor{}, err
	}
	method := typeinfo.MethodDescriptor{ID: m.id(fn.ID), Name: typeinfo.KnownFact(fn.Name, m.scope, ev), Signature: typeinfo.KnownFact(fn.Signature, m.scope, ev), Visibility: typeinfo.KnownFact(goVisibility(fn.Visibility), m.scope, ev), Modifiers: typeinfo.KnownFact([]typeinfo.Modifier{}, m.scope, ev), Parameters: []typeinfo.ParameterDescriptor{}, Returns: []typeinfo.ReturnDescriptor{}, GenericParameters: []typeinfo.GenericParameter{}, Attributes: []typeinfo.AttributeUse{}, Declaration: decl}
	if fn.Receiver != nil {
		receiver := m.parameter(fn.ID+":receiver", 0, *fn.Receiver, ev)
		method.Receiver = &receiver
	}
	for i, p := range fn.Parameters {
		method.Parameters = append(method.Parameters, m.parameter(fn.ID, i, p, ev))
	}
	for i, p := range fn.Returns {
		method.Returns = append(method.Returns, typeinfo.ReturnDescriptor{ID: m.id(fmt.Sprintf("%s:return:%d", fn.ID, i)), Name: typeinfo.KnownFact(p.Name, m.scope, ev), Position: typeinfo.KnownFact(i, m.scope, ev), Type: m.typeReference(p.Type, ev), Attributes: []typeinfo.AttributeUse{}})
	}
	for i, p := range fn.TypeParameters {
		method.GenericParameters = append(method.GenericParameters, m.generic(fn.ID, i, p, ev))
	}
	return method, nil
}
func (m *GoMapper) constant(path string, c analysis.ConstantInfo) (typeinfo.EnumMember, error) {
	ev, err := m.evidence(path, c.ID, "go-ast-syntax", c.Range)
	if err != nil {
		return typeinfo.EnumMember{}, err
	}
	decl, err := m.occurrence(path, c.ID+":declaration", "definition", c.ID, c.Range, true)
	if err != nil {
		return typeinfo.EnumMember{}, err
	}
	value := typeinfo.UnresolvedFact[string](m.scope, "constant expression has not been resolved by the Go type checker")
	typeName := nonemptyName(c.ResolvedType, c.DeclaredType)
	if c.Resolution == "go-types" {
		checked := ev
		checked.ID += "-checked"
		checked.Producer = "go-types"
		value = typeinfo.KnownFact(c.Value, m.scope, checked)
	}
	typeRef := m.typeReference(analysis.TypeExpression{Text: typeName}, ev)
	if c.ResolvedType != "" && c.Resolution == "go-types" {
		checked := ev
		checked.ID += "-checked"
		checked.Producer = "go-types"
		typeRef.Name = typeinfo.KnownFact(typeName, m.scope, checked)
	}
	if typeName == "" {
		typeRef.Name = typeinfo.UnresolvedFact[string](m.scope, "constant type unresolved")
	}
	member := typeinfo.EnumMember{ID: m.id(c.ID), GroupID: m.id(c.GroupID), Name: typeinfo.KnownFact(c.Name, m.scope, ev), Expression: typeinfo.KnownFact(c.Expression, m.scope, ev), Value: typeinfo.ConstantValue{Type: typeRef, Value: value}, Declaration: decl, Attributes: []typeinfo.AttributeUse{}, Occurrences: []typeinfo.OccurrenceLink{}}
	for _, o := range c.References {
		id := fmt.Sprintf("%s:use:%d:%d", o.SymbolID, o.Range.Start, o.Range.End)
		occurrence, err := m.occurrence(path, id, o.Role, o.SymbolID, o.Range, o.Resolution == "go-types")
		if err != nil {
			return typeinfo.EnumMember{}, err
		}
		member.Occurrences = append(member.Occurrences, *occurrence)
	}
	return member, nil
}
