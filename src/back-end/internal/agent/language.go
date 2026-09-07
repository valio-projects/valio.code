package agent

import (
	"path/filepath"
	"strings"
)

func language(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cs":
		return "csharp"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx", ".hxx", ".hh":
		return "cpp"
	case ".md":
		return "markdown"
	default:
		return "text"
	}
}
