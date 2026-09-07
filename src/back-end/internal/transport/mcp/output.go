package mcp

// Output preserves typed application results without substituting natural-language claims.
type Output struct {
	// Data is the typed application result returned by a tool.
	Data any `json:"data"`
}
