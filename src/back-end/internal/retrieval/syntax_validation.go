package retrieval

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/domain"
	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

func validateSyntaxSource(source Source, report syntaxfacts.Report) error {
	if source.ID == "" || source.RepositoryID == "" || source.Path == "" || source.Language == "" {
		return fmt.Errorf("syntax retrieval source requires id, repository, path, and language")
	}
	if err := domain.ValidateRelativePath(source.Path, false); err != nil {
		return fmt.Errorf("syntax retrieval source path is invalid")
	}
	if !utf8.ValidString(source.Content) {
		return fmt.Errorf("syntax retrieval source must contain valid UTF-8")
	}
	if strings.EqualFold(source.Language, "go") || report.Language != source.Language || report.Path != source.Path {
		return fmt.Errorf("syntax report does not match source")
	}
	if err := report.Validate(source.Content); err != nil {
		return fmt.Errorf("invalid syntax report")
	}
	return nil
}

func validateSyntaxDeclarations(content string, values []syntaxfacts.Declaration) ([]syntaxfacts.Declaration, map[string][]syntaxfacts.Declaration, error) {
	byID := make(map[string]syntaxfacts.Declaration, len(values))
	for _, value := range values {
		if value.ID == "" || strings.TrimSpace(value.ID) != value.ID || !validSyntaxChunkKind(syntaxChunkKind(value.Kind)) || value.Start < 0 || value.End <= value.Start || value.End > len(content) || !utf8Boundary(content, value.Start) || !utf8Boundary(content, value.End) {
			return nil, nil, fmt.Errorf("invalid syntax declaration")
		}
		if _, exists := byID[value.ID]; exists {
			return nil, nil, fmt.Errorf("duplicate syntax declaration id")
		}
		byID[value.ID] = value
	}
	for _, value := range values {
		if value.ParentID == "" {
			continue
		}
		parent, exists := byID[value.ParentID]
		if !exists || parent.Start > value.Start || value.End > parent.End || parent.ID == value.ID {
			return nil, nil, fmt.Errorf("invalid syntax declaration parent")
		}
		seen := map[string]bool{value.ID: true}
		for current := parent; current.ParentID != ""; current = byID[current.ParentID] {
			if seen[current.ID] {
				return nil, nil, fmt.Errorf("syntax declaration parent cycle")
			}
			seen[current.ID] = true
			if _, exists := byID[current.ParentID]; !exists {
				return nil, nil, fmt.Errorf("invalid syntax declaration parent")
			}
		}
	}
	depths := map[string]int{}
	for _, value := range values {
		for current := value; current.ParentID != ""; current = byID[current.ParentID] {
			depths[value.ID]++
		}
	}
	sorted := append([]syntaxfacts.Declaration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		if sorted[i].End != sorted[j].End {
			return sorted[i].End > sorted[j].End
		}
		if depths[sorted[i].ID] != depths[sorted[j].ID] {
			return depths[sorted[i].ID] < depths[sorted[j].ID]
		}
		return sorted[i].ID < sorted[j].ID
	})
	stack := []syntaxfacts.Declaration{}
	for _, value := range sorted {
		for len(stack) > 0 && stack[len(stack)-1].End <= value.Start {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			container := stack[len(stack)-1]
			if value.End > container.End || !syntaxAncestor(container.ID, value.ID, byID) {
				return nil, nil, fmt.Errorf("overlapping syntax declarations")
			}
		}
		stack = append(stack, value)
	}
	children := map[string][]syntaxfacts.Declaration{}
	for _, value := range sorted {
		if value.ParentID != "" {
			children[value.ParentID] = append(children[value.ParentID], value)
		}
	}
	for id := range children {
		sort.Slice(children[id], func(i, j int) bool {
			if children[id][i].Start != children[id][j].Start {
				return children[id][i].Start < children[id][j].Start
			}
			return children[id][i].ID < children[id][j].ID
		})
	}
	return sorted, children, nil
}

func syntaxAncestor(ancestor, id string, values map[string]syntaxfacts.Declaration) bool {
	for current := values[id]; current.ParentID != ""; current = values[current.ParentID] {
		if current.ParentID == ancestor {
			return true
		}
	}
	return false
}

func utf8Boundary(content string, offset int) bool {
	return offset == 0 || offset == len(content) || (offset > 0 && offset < len(content) && utf8.RuneStart(content[offset]))
}

func syntaxChunkKind(kind string) domainretrieval.ChunkKind {
	switch kind {
	case "class", "interface", "struct", "enum", "enum_member":
		return domainretrieval.ChunkType
	case "function":
		return domainretrieval.ChunkFunction
	case "method":
		return domainretrieval.ChunkMethod
	case "field", "property":
		return domainretrieval.ChunkStatement
	default:
		return ""
	}
}

func validSyntaxChunkKind(kind domainretrieval.ChunkKind) bool {
	return kind == domainretrieval.ChunkType || kind == domainretrieval.ChunkFunction || kind == domainretrieval.ChunkMethod || kind == domainretrieval.ChunkStatement
}
