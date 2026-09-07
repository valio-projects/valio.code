package syntax

import (
	"encoding/json"
	"errors"
)

func validateReport(payload json.RawMessage, path, language string, size int) error {
	invalid := errors.New("INVALID_SYNTAX_REPORT")
	var report map[string]any
	if json.Unmarshal(payload, &report) != nil || report["schema"] != float64(1) || report["path"] != path || report["language"] != language {
		return invalid
	}
	if _, ok := report["validSyntax"].(bool); !ok {
		return invalid
	}
	capabilities, ok := report["capabilities"].(map[string]any)
	if !ok || capabilities["syntax"] != true || capabilities["semanticResolution"] != false {
		return errors.New("SYNTAX_CAPABILITY_UNAVAILABLE")
	}
	for _, name := range []string{"symbols", "references", "imports", "diagnostics"} {
		if _, ok := report[name].([]any); !ok {
			return invalid
		}
	}
	count := 0
	var check func(any, int) bool
	check = func(value any, depth int) bool {
		count++
		if depth > 128 || count > 200000 {
			return false
		}
		switch v := value.(type) {
		case map[string]any:
			start, hasStart := v["start"]
			end, hasEnd := v["end"]
			if hasStart != hasEnd {
				return false
			}
			if hasStart {
				a, okA := start.(float64)
				b, okB := end.(float64)
				if !okA || !okB || a < 0 || b < a || b > float64(size) || a != float64(int(a)) || b != float64(int(b)) {
					return false
				}
			}
			if resolution, ok := v["resolution"]; ok && resolution != "unresolved" {
				return false
			}
			if resolved, ok := v["semanticResolution"]; ok && resolved != false {
				return false
			}
			for _, child := range v {
				if !check(child, depth+1) {
					return false
				}
			}
		case []any:
			for _, child := range v {
				if !check(child, depth+1) {
					return false
				}
			}
		}
		return true
	}
	if !check(report, 0) {
		return invalid
	}
	return nil
}
