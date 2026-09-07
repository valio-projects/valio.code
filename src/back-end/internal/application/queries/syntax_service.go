package queries

import (
	"context"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/application/fault"
)

// Syntax returns matching declarations plus their descendant members. Reference
// observations retain unresolved status; no cross-file target is guessed.
func (s Service) Syntax(ctx context.Context, q SyntaxQuery) (SyntaxResult, error) {
	result := SyntaxResult{Reports: []SyntaxFile{}, Status: "unsupported"}
	if q.Name == "" && q.FileID == "" || len(q.Name) > 4096 {
		return result, fault.ErrInvalid
	}
	v, _, err := s.scopedChunks(ctx, q.Scope)
	if err != nil {
		return result, err
	}
	result.ViewID = v.ID
	artifacts, err := s.Store.Artifacts(ctx, v)
	if err != nil {
		return result, err
	}
	for _, f := range v.Files {
		if q.FileID != "" && f.ID != q.FileID || len(q.Scope.ProjectIDs) > 0 && !overlaps(f.ProjectIDs, q.Scope.ProjectIDs) {
			continue
		}
		for _, a := range artifacts {
			if a.FileID != f.ID || len(a.Syntax) == 0 {
				continue
			}
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			result.Status = "partial"
			var report map[string]json.RawMessage
			if json.Unmarshal(a.Syntax, &report) != nil {
				return result, fault.ErrInvalid
			}
			if q.Name != "" {
				var symbols []map[string]json.RawMessage
				if json.Unmarshal(report["symbols"], &symbols) != nil {
					return result, fault.ErrInvalid
				}
				selected := map[string]bool{}
				for _, symbol := range symbols {
					if jsonString(symbol["name"]) == q.Name {
						selected[jsonString(symbol["id"])] = true
					}
				}
				if len(selected) == 0 {
					continue
				}
				for pass := 0; pass < 32; pass++ {
					changed := false
					for _, symbol := range symbols {
						if selected[jsonString(symbol["parentId"])] && !selected[jsonString(symbol["id"])] {
							selected[jsonString(symbol["id"])] = true
							changed = true
						}
					}
					if !changed {
						break
					}
				}
				picked := []map[string]json.RawMessage{}
				for _, symbol := range symbols {
					if selected[jsonString(symbol["id"])] {
						picked = append(picked, symbol)
					}
				}
				report["symbols"], _ = json.Marshal(picked)
				// Identifier equality is an unresolved candidate-use filter, not resolution.
				var refs []map[string]json.RawMessage
				if json.Unmarshal(report["references"], &refs) != nil {
					return result, fault.ErrInvalid
				}
				pickedRefs := []map[string]json.RawMessage{}
				for _, ref := range refs {
					if jsonString(ref["name"]) == q.Name {
						pickedRefs = append(pickedRefs, ref)
					}
				}
				report["references"], _ = json.Marshal(pickedRefs)
			}
			raw, err := json.Marshal(report)
			if err != nil {
				return result, err
			}
			result.Reports = append(result.Reports, SyntaxFile{FileID: f.ID, RepositoryID: string(f.RepositoryID), Path: f.Path, Language: f.Language, Report: raw})
		}
	}
	return result, nil
}

func jsonString(raw json.RawMessage) string {
	var value string
	_ = json.Unmarshal(raw, &value)
	return value
}
