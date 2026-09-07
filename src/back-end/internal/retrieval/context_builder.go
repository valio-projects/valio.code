package retrieval

import (
	"fmt"
	"sort"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

// ContextBuilder expands one exact caller-selected root with its parent and
// direct children. It never fetches or discovers additional chunks.
type ContextBuilder struct{}

// Build creates cited context for rootID from callerChunks. rootID must occur
// exactly in callerChunks. The method applies repository, file, project, range,
// and byte-budget checks before adding an item.
func (ContextBuilder) Build(rootID string, callerChunks []domainretrieval.Chunk, options ContextOptions) (Context, error) {
	if rootID == "" {
		return Context{}, fmt.Errorf("context root is required")
	}
	budget, err := contextBudget(options.MaxBytes)
	if err != nil {
		return Context{}, err
	}
	byID := map[string]domainretrieval.Chunk{}
	for _, chunk := range callerChunks {
		if chunk.ID == "" || byID[chunk.ID].ID != "" {
			return Context{}, fmt.Errorf("caller chunks require unique nonempty IDs")
		}
		byID[chunk.ID] = chunk
	}
	root, found := byID[rootID]
	if !found {
		return Context{}, fmt.Errorf("context root is not in caller chunks")
	}
	children := map[string][]domainretrieval.Chunk{}
	for _, chunk := range callerChunks {
		if chunk.ParentID != "" {
			children[chunk.ParentID] = append(children[chunk.ParentID], chunk)
		}
	}
	for parent := range children {
		sort.Slice(children[parent], func(i, j int) bool {
			if children[parent][i].Start != children[parent][j].Start {
				return children[parent][i].Start < children[parent][j].Start
			}
			return children[parent][i].ID < children[parent][j].ID
		})
	}
	candidates := []domainretrieval.Chunk{root}
	if root.ParentID != "" {
		if parent, ok := byID[root.ParentID]; ok {
			candidates = append(candidates, parent)
		}
	}
	candidates = append(candidates, children[root.ID]...)
	context := Context{ViewID: options.ViewID, Items: []ContextItem{}, Omitted: []ContextOmission{}}
	selected := map[string]bool{}
	ranges := []domainretrieval.Chunk{}
	for _, candidate := range candidates {
		if selected[candidate.ID] {
			continue
		}
		selected[candidate.ID] = true
		if !sameScope(root, candidate, options) {
			context.Omitted = append(context.Omitted, ContextOmission{ChunkID: candidate.ID, Reason: OmissionProjectScope})
			continue
		}
		if overlaps(ranges, candidate) {
			context.Omitted = append(context.Omitted, ContextOmission{ChunkID: candidate.ID, Reason: OmissionDuplicateRange})
			continue
		}
		item := ContextItem{ChunkID: candidate.ID, FileID: candidate.FileID, Start: candidate.Start, End: candidate.End, Text: candidate.Text, Header: candidate.Header, ParentID: candidate.ParentID, Profile: options.Profile}
		size := contextItemBytes(item)
		if context.ByteCount+size > budget {
			context.Omitted = append(context.Omitted, ContextOmission{ChunkID: candidate.ID, Reason: OmissionBudget})
			continue
		}
		context.Items = append(context.Items, item)
		context.ByteCount += size
		if candidate.Text != "" {
			ranges = append(ranges, candidate)
		}
	}
	return context, nil
}

func contextBudget(value int) (int, error) {
	if value == 0 {
		return 8000, nil
	}
	if value < 1 || value > 8000 {
		return 0, fmt.Errorf("context byte budget must be between 1 and 8000")
	}
	return value, nil
}

func sameScope(root, candidate domainretrieval.Chunk, options ContextOptions) bool {
	if root.FileID != candidate.FileID || root.RepositoryID != candidate.RepositoryID {
		return false
	}
	allowed := root.ProjectIDs
	if len(options.ProjectIDs) > 0 {
		allowed = options.ProjectIDs
	}
	if len(allowed) == 0 {
		return len(candidate.ProjectIDs) == 0
	}
	for _, candidateProject := range candidate.ProjectIDs {
		for _, project := range allowed {
			if candidateProject == project && contains(root.ProjectIDs, project) {
				return true
			}
		}
	}
	return false
}

func overlaps(selected []domainretrieval.Chunk, candidate domainretrieval.Chunk) bool {
	if candidate.Text == "" {
		return false
	}
	for _, item := range selected {
		if item.FileID == candidate.FileID && item.Text != "" && candidate.Start < item.End && item.Start < candidate.End {
			return true
		}
	}
	return false
}

func contextItemBytes(item ContextItem) int {
	return len(item.ChunkID) + len(item.FileID) + len(item.ParentID) + len(item.Header) + len(item.Text) + len(item.Profile)
}
