package snapshots

import (
	"context"
	"encoding/json"
)

// SyntaxAnalyzer supplies bounded syntax evidence from a configured language
// helper. Implementations must never execute the analyzed project or its build.
type SyntaxAnalyzer interface {
	Profile() string
	Analyze(context.Context, string, string, string) (json.RawMessage, error)
}
