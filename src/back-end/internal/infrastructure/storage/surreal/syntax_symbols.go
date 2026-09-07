package surreal

import (
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/search"
)

// syntaxSymbols maps written declarations to source-search metadata without
// interpreting parser observations as compiler-resolved symbols.
func syntaxSymbols(raw json.RawMessage) ([]search.Symbol, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var report struct {
		Symbols []struct {
			Name  string `json:"name"`
			Kind  string `json:"kind"`
			Start int    `json:"start"`
			End   int    `json:"end"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, err
	}
	result := []search.Symbol{}
	for _, s := range report.Symbols {
		result = append(result, search.Symbol{Name: s.Name, Kind: s.Kind, Start: s.Start, End: s.End})
	}
	return result, nil
}
