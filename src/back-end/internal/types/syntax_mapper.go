package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// SyntaxMapper translates parser-only declarations into evidence-backed type
// descriptors. It never promotes a written name to a semantic symbol target.
type SyntaxMapper struct {
	scope      typeinfo.TypeScope
	repository domain.RepositoryID
	source     string
	lineStarts []int
}

// NewSyntaxMapper validates immutable context shared by every mapped fact.
func NewSyntaxMapper(scope typeinfo.TypeScope, repository domain.RepositoryID, source string) (*SyntaxMapper, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	if repository == "" {
		return nil, fmt.Errorf("syntax mapper requires repository identity")
	}
	starts := []int{0}
	for index := 0; index < len(source); index++ {
		if source[index] == '\n' {
			starts = append(starts, index+1)
		}
	}
	return &SyntaxMapper{scope: scope, repository: repository, source: source, lineStarts: starts}, nil
}

// FromSyntaxReport maps one syntax-helper report without assuming compiler or
// runtime information. All resulting semantic references remain unresolved.
func FromSyntaxReport(scope typeinfo.TypeScope, repositoryID domain.RepositoryID, source string, report syntaxfacts.Report) ([]typeinfo.TypeDescriptor, error) {
	mapper, err := NewSyntaxMapper(scope, repositoryID, source)
	if err != nil {
		return nil, err
	}
	return mapper.Map(report)
}

// Map converts direct type declarations and their directly-parented members.
func (m *SyntaxMapper) Map(report syntaxfacts.Report) ([]typeinfo.TypeDescriptor, error) {
	if err := report.Validate(m.source); err != nil {
		return nil, err
	}
	declarations := make(map[string]syntaxfacts.Declaration, len(report.Symbols))
	for _, declaration := range report.Symbols {
		if declaration.ID == "" {
			return nil, fmt.Errorf("syntax declaration %q has no identity", declaration.Name)
		}
		if _, exists := declarations[declaration.ID]; exists {
			return nil, fmt.Errorf("duplicate syntax declaration identity %q", declaration.ID)
		}
		if _, err := m.sourceRange(report.Path, declaration.Start, declaration.End); err != nil {
			return nil, fmt.Errorf("syntax declaration %q: %w", declaration.ID, err)
		}
		declarations[declaration.ID] = declaration
	}

	types := make(map[string]*typeinfo.TypeDescriptor)
	order := []string{}
	for _, declaration := range report.Symbols {
		kind, ok := syntaxTypeKind(declaration.Kind)
		if !ok {
			continue
		}
		if strings.TrimSpace(declaration.Name) == "" {
			// TypeDescriptor requires a catalogable nonempty name. The caller
			// retains this anonymous syntax node in its syntax graph.
			continue
		}
		descriptor, err := m.typeDescriptor(report, declaration, kind)
		if err != nil {
			return nil, err
		}
		types[declaration.ID] = &descriptor
		order = append(order, declaration.ID)
	}

	for _, declaration := range report.Symbols {
		parent, exists := types[declaration.ParentID]
		if !exists {
			continue
		}
		if err := m.addMember(report, parent, declaration); err != nil {
			return nil, err
		}
	}
	result := make([]typeinfo.TypeDescriptor, 0, len(order))
	for _, id := range order {
		descriptor := *types[id]
		if err := descriptor.Validate(); err != nil {
			return nil, fmt.Errorf("mapped syntax type %q: %w", id, err)
		}
		result = append(result, descriptor)
	}
	return result, nil
}

func syntaxTypeKind(kind string) (typeinfo.TypeKind, bool) {
	switch kind {
	case "class":
		return typeinfo.TypeClass, true
	case "struct":
		return typeinfo.TypeStruct, true
	case "interface":
		return typeinfo.TypeInterface, true
	case "enum":
		return typeinfo.TypeEnum, true
	default:
		return "", false
	}
}

func (m *SyntaxMapper) id(local string) string {
	sum := sha256.Sum256([]byte(string(m.repository) + "\x00" + local))
	return hex.EncodeToString(sum[:16])
}

func (m *SyntaxMapper) sourceRange(path string, start, end int) (domain.SourceRange, error) {
	if start < 0 || end < start || end > len(m.source) {
		return domain.SourceRange{}, fmt.Errorf("syntax range [%d,%d) is outside supplied source", start, end)
	}
	position := func(offset int) domain.Position {
		line := sort.Search(len(m.lineStarts), func(index int) bool { return m.lineStarts[index] > offset }) - 1
		return domain.Position{Line: line, Character: offset - m.lineStarts[line]}
	}
	return domain.SourceRange{RepositoryID: m.repository, Path: path, Encoding: domain.EncodingUTF8, Start: position(start), End: position(end)}, nil
}

func (m *SyntaxMapper) evidence(path, local string, start, end int) (typeinfo.EvidenceRef, error) {
	rangeValue, err := m.sourceRange(path, start, end)
	if err != nil {
		return typeinfo.EvidenceRef{}, err
	}
	return typeinfo.EvidenceRef{ID: m.id(path + ":" + local + ":syntax"), Producer: "syntax-wasm", Range: &rangeValue}, nil
}

func (m *SyntaxMapper) unresolvedReference() typeinfo.SymbolReference {
	return typeinfo.SymbolReference{Resolution: typeinfo.ReferenceUnresolved, Reason: "syntax-only report has no semantic resolution", Candidates: []typeinfo.SymbolIdentity{}}
}

func (m *SyntaxMapper) typeReference(written string, evidence typeinfo.EvidenceRef) typeinfo.TypeReference {
	if strings.TrimSpace(written) == "" {
		return typeinfo.TypeReference{Name: typeinfo.UnresolvedFact[string](m.scope, "written type is unavailable in syntax report"), Symbol: m.unresolvedReference(), Attributes: []typeinfo.AttributeUse{}}
	}
	return typeinfo.TypeReference{Name: typeinfo.KnownFact(written, m.scope, evidence), Symbol: m.unresolvedReference(), Attributes: []typeinfo.AttributeUse{}}
}

func (m *SyntaxMapper) occurrence(path, local string, start, end int, evidence typeinfo.EvidenceRef) (*typeinfo.OccurrenceLink, error) {
	rangeValue, err := m.sourceRange(path, start, end)
	if err != nil {
		return nil, err
	}
	return &typeinfo.OccurrenceLink{ID: m.id(path + ":" + local + ":declaration"), Range: rangeValue, Role: typeinfo.KnownFact("definition", m.scope, evidence), Symbol: m.unresolvedReference()}, nil
}
