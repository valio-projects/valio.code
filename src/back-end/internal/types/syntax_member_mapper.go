package types

import (
	"fmt"
	"strings"

	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// addMember attaches only a declaration whose direct lexical parent is the type.
func (m *SyntaxMapper) addMember(report syntaxfacts.Report, parent *typeinfo.TypeDescriptor, declaration syntaxfacts.Declaration) error {
	switch declaration.Kind {
	case "field":
		field, err := m.field(report, declaration)
		if err != nil {
			return err
		}
		parent.Fields = append(parent.Fields, field)
	case "property":
		property, err := m.property(report, declaration)
		if err != nil {
			return err
		}
		parent.Properties = append(parent.Properties, property)
	case "method":
		method, err := m.method(report, declaration)
		if err != nil {
			return err
		}
		parent.Methods = append(parent.Methods, method)
	case "enum_member":
		if parent.Enum == nil {
			return nil
		}
		member, err := m.enumMember(report, parent, declaration)
		if err != nil {
			return err
		}
		parent.Enum.Members = append(parent.Enum.Members, member)
	}
	return nil
}

func (m *SyntaxMapper) field(report syntaxfacts.Report, declaration syntaxfacts.Declaration) (typeinfo.FieldDescriptor, error) {
	evidence, err := m.evidence(report.Path, declaration.ID, declaration.Start, declaration.End)
	if err != nil {
		return typeinfo.FieldDescriptor{}, err
	}
	link, err := m.occurrence(report.Path, declaration.ID, declaration.Start, declaration.End, evidence)
	if err != nil {
		return typeinfo.FieldDescriptor{}, err
	}
	return typeinfo.FieldDescriptor{
		ID:          m.id(report.Path + ":" + declaration.ID),
		Name:        m.nameFact(declaration.Name, evidence, "field name is unavailable in syntax report"),
		Type:        m.typeReference(stringValue(declaration.Type), evidence),
		Visibility:  typeinfo.KnownFact(syntaxVisibility(declaration.Visibility), m.scope, evidence),
		Modifiers:   typeinfo.KnownFact(syntaxModifiers(declaration.Modifiers), m.scope, evidence),
		Attributes:  m.attributes(report.Path, declaration.ID, declaration.Attributes, declaration.Start, declaration.End, evidence),
		Tag:         typeinfo.UnresolvedFact[string](m.scope, "syntax report does not classify declaration text as a Go struct tag"),
		Declaration: link,
	}, nil
}

func (m *SyntaxMapper) property(report syntaxfacts.Report, declaration syntaxfacts.Declaration) (typeinfo.PropertyDescriptor, error) {
	evidence, err := m.evidence(report.Path, declaration.ID, declaration.Start, declaration.End)
	if err != nil {
		return typeinfo.PropertyDescriptor{}, err
	}
	link, err := m.occurrence(report.Path, declaration.ID, declaration.Start, declaration.End, evidence)
	if err != nil {
		return typeinfo.PropertyDescriptor{}, err
	}
	parameters, err := m.parameters(report, declaration)
	if err != nil {
		return typeinfo.PropertyDescriptor{}, err
	}
	return typeinfo.PropertyDescriptor{
		ID:          m.id(report.Path + ":" + declaration.ID),
		Name:        m.nameFact(declaration.Name, evidence, "property name is unavailable in syntax report"),
		Type:        m.typeReference(stringValue(declaration.Type), evidence),
		Visibility:  typeinfo.KnownFact(syntaxVisibility(declaration.Visibility), m.scope, evidence),
		Modifiers:   typeinfo.KnownFact(syntaxModifiers(declaration.Modifiers), m.scope, evidence),
		Parameters:  parameters,
		Attributes:  m.attributes(report.Path, declaration.ID, declaration.Attributes, declaration.Start, declaration.End, evidence),
		Declaration: link,
	}, nil
}

func (m *SyntaxMapper) method(report syntaxfacts.Report, declaration syntaxfacts.Declaration) (typeinfo.MethodDescriptor, error) {
	evidence, err := m.evidence(report.Path, declaration.ID, declaration.Start, declaration.End)
	if err != nil {
		return typeinfo.MethodDescriptor{}, err
	}
	link, err := m.occurrence(report.Path, declaration.ID, declaration.Start, declaration.End, evidence)
	if err != nil {
		return typeinfo.MethodDescriptor{}, err
	}
	parameters, err := m.parameters(report, declaration)
	if err != nil {
		return typeinfo.MethodDescriptor{}, err
	}
	returns := []typeinfo.ReturnDescriptor{}
	if declaration.Type != nil && strings.TrimSpace(*declaration.Type) != "" {
		returns = append(returns, typeinfo.ReturnDescriptor{ID: m.id(fmt.Sprintf("%s:%s:return:0", report.Path, declaration.ID)), Name: typeinfo.UnresolvedFact[string](m.scope, "syntax report has no named return binding"), Position: typeinfo.KnownFact(0, m.scope, evidence), Type: m.typeReference(*declaration.Type, evidence), Attributes: []typeinfo.AttributeUse{}})
	}
	return typeinfo.MethodDescriptor{
		ID:                m.id(report.Path + ":" + declaration.ID),
		Name:              m.nameFact(declaration.Name, evidence, "method name is unavailable in syntax report"),
		Signature:         typeinfo.UnresolvedFact[string](m.scope, "syntax report does not provide a canonical method signature"),
		Visibility:        typeinfo.KnownFact(syntaxVisibility(declaration.Visibility), m.scope, evidence),
		Modifiers:         typeinfo.KnownFact(syntaxModifiers(declaration.Modifiers), m.scope, evidence),
		Parameters:        parameters,
		Returns:           returns,
		GenericParameters: []typeinfo.GenericParameter{},
		Attributes:        m.attributes(report.Path, declaration.ID, declaration.Attributes, declaration.Start, declaration.End, evidence),
		Declaration:       link,
	}, nil
}

func (m *SyntaxMapper) parameters(report syntaxfacts.Report, declaration syntaxfacts.Declaration) ([]typeinfo.ParameterDescriptor, error) {
	parameters := make([]typeinfo.ParameterDescriptor, 0, len(declaration.Parameters))
	for index, parameter := range declaration.Parameters {
		parameterEvidence, err := m.evidence(report.Path, fmt.Sprintf("%s:parameter:%d", declaration.ID, index), parameter.Start, parameter.End)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, typeinfo.ParameterDescriptor{
			ID:           m.id(fmt.Sprintf("%s:%s:parameter:%d", report.Path, declaration.ID, index)),
			Name:         m.nameFact(parameter.Name, parameterEvidence, "parameter name is unavailable in syntax report"),
			Position:     typeinfo.KnownFact(index, m.scope, parameterEvidence),
			Type:         m.typeReference(stringValue(parameter.Type), parameterEvidence),
			Modifiers:    typeinfo.KnownFact(syntaxModifiers(parameter.Modifiers), m.scope, parameterEvidence),
			DefaultValue: typeinfo.UnresolvedFact[string](m.scope, "syntax report does not collect default parameter values"),
			Attributes:   m.attributes(report.Path, fmt.Sprintf("%s:parameter:%d", declaration.ID, index), parameter.Attributes, parameter.Start, parameter.End, parameterEvidence),
		})
	}
	return parameters, nil
}

func (m *SyntaxMapper) enumMember(report syntaxfacts.Report, parent *typeinfo.TypeDescriptor, declaration syntaxfacts.Declaration) (typeinfo.EnumMember, error) {
	evidence, err := m.evidence(report.Path, declaration.ID, declaration.Start, declaration.End)
	if err != nil {
		return typeinfo.EnumMember{}, err
	}
	link, err := m.occurrence(report.Path, declaration.ID, declaration.Start, declaration.End, evidence)
	if err != nil {
		return typeinfo.EnumMember{}, err
	}
	value := typeinfo.UnresolvedFact[string](m.scope, "syntax report does not evaluate enum constant expressions")
	expression := typeinfo.UnresolvedFact[string](m.scope, "enum value expression is not present in syntax report")
	if declaration.EnumValue != nil {
		expression = typeinfo.KnownFact(*declaration.EnumValue, m.scope, evidence)
	}
	return typeinfo.EnumMember{
		ID:          m.id(report.Path + ":" + declaration.ID),
		GroupID:     parent.ID,
		Name:        m.nameFact(declaration.Name, evidence, "enum member name is unavailable in syntax report"),
		Expression:  expression,
		Value:       typeinfo.ConstantValue{Type: m.typeReference(*parent.Name.Value, evidence), Value: value},
		Attributes:  m.attributes(report.Path, declaration.ID, declaration.Attributes, declaration.Start, declaration.End, evidence),
		Declaration: link,
		Occurrences: []typeinfo.OccurrenceLink{},
	}, nil
}

func (m *SyntaxMapper) nameFact(name string, evidence typeinfo.EvidenceRef, reason string) typeinfo.Fact[string] {
	if strings.TrimSpace(name) == "" {
		return typeinfo.UnresolvedFact[string](m.scope, reason)
	}
	return typeinfo.KnownFact(name, m.scope, evidence)
}
