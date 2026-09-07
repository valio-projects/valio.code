package types

import (
	"strconv"
	"strings"

	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// typeDescriptor maps only facts directly observed in a type declaration.
func (m *SyntaxMapper) typeDescriptor(report syntaxfacts.Report, declaration syntaxfacts.Declaration, kind typeinfo.TypeKind) (typeinfo.TypeDescriptor, error) {
	evidence, err := m.evidence(report.Path, declaration.ID, declaration.Start, declaration.End)
	if err != nil {
		return typeinfo.TypeDescriptor{}, err
	}
	declarationLink, err := m.occurrence(report.Path, declaration.ID, declaration.Start, declaration.End, evidence)
	if err != nil {
		return typeinfo.TypeDescriptor{}, err
	}
	descriptor := typeinfo.TypeDescriptor{
		ID:                 m.id(report.Path + ":" + declaration.ID),
		Scope:              m.scope,
		Name:               typeinfo.KnownFact(declaration.Name, m.scope, evidence),
		FullyQualifiedName: typeinfo.UnresolvedFact[string](m.scope, "syntax report does not establish a qualified semantic name"),
		Language:           typeinfo.KnownFact(report.Language, m.scope, evidence),
		Kind:               typeinfo.KnownFact(kind, m.scope, evidence),
		Visibility:         typeinfo.KnownFact(syntaxVisibility(declaration.Visibility), m.scope, evidence),
		Modifiers:          typeinfo.KnownFact(syntaxModifiers(declaration.Modifiers), m.scope, evidence),
		Symbol:             m.unresolvedReference(),
		Declaration:        declarationLink,
		Fields:             []typeinfo.FieldDescriptor{},
		Properties:         []typeinfo.PropertyDescriptor{},
		Methods:            []typeinfo.MethodDescriptor{},
		Constructors:       []typeinfo.MethodDescriptor{},
		GenericParameters:  []typeinfo.GenericParameter{},
		Attributes:         m.attributes(report.Path, declaration.ID, declaration.Attributes, declaration.Start, declaration.End, evidence),
		Constants:          []typeinfo.EnumMember{},
		EmbeddedTypes:      []typeinfo.TypeReference{},
		Layout:             typeinfo.UnresolvedLayout(m.scope, "syntax-only analysis cannot establish physical layout"),
	}
	if kind == typeinfo.TypeEnum {
		descriptor.Enum = &typeinfo.EnumDescriptor{UnderlyingType: m.typeReference(stringValue(declaration.UnderlyingType), evidence), Members: []typeinfo.EnumMember{}}
	}
	return descriptor, nil
}

func syntaxVisibility(value string) typeinfo.Visibility {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "public":
		return typeinfo.VisibilityPublic
	case "private":
		return typeinfo.VisibilityPrivate
	case "protected":
		return typeinfo.VisibilityProtected
	case "internal":
		return typeinfo.VisibilityInternal
	case "protected internal":
		return typeinfo.VisibilityProtectedInternal
	case "private protected":
		return typeinfo.VisibilityPrivateProtected
	case "package":
		return typeinfo.VisibilityPackage
	case "module":
		return typeinfo.VisibilityModule
	default:
		return typeinfo.VisibilityUnknown
	}
}

func syntaxModifiers(values []string) []typeinfo.Modifier {
	result := make([]typeinfo.Modifier, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, typeinfo.Modifier(trimmed))
		}
	}
	return result
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (m *SyntaxMapper) attributes(path, parent string, values []string, start, end int, evidence typeinfo.EvidenceRef) []typeinfo.AttributeUse {
	attributes := make([]typeinfo.AttributeUse, 0, len(values))
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		local := attributeLocalID(parent, index)
		occurrence, err := m.occurrence(path, local, start, end, evidence)
		if err != nil {
			continue
		}
		raw := typeinfo.KnownFact(value, m.scope, evidence)
		attributes = append(attributes, typeinfo.AttributeUse{
			ID:                  m.id(path + ":" + local),
			RawSyntax:           &raw,
			Name:                typeinfo.UnresolvedFact[string](m.scope, "syntax report retained a complete attribute list without parsing its name"),
			AttributeClass:      m.unresolvedReference(),
			PositionalArguments: []typeinfo.AttributeValue{},
			NamedArguments:      []typeinfo.NamedAttributeArgument{},
			Occurrence:          occurrence,
		})
	}
	return attributes
}

func attributeLocalID(parent string, index int) string {
	return parent + ":attribute:" + strconv.Itoa(index)
}
