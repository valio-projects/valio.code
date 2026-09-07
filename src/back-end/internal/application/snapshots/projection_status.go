package snapshots

// projectionStatus advertises only producers run by this publication profile.
// Partial compiler/reference data never promises complete build semantics.
func projectionStatus() map[string]string {
	return map[string]string{
		"source_text": "ready", "symbols": "partial", "types": "partial",
		"compiler": "partial", "references": "partial", "calls": "partial",
		"reads_writes":     "partial",
		"retrieval_chunks": "ready", "lexical": "ready", "structural_fingerprint": "partial",
		"configuration_graph": "unsupported", "git_diff": "unsupported",
		"cfg": "unsupported", "dataflow": "unsupported", "vectors": "not_requested",
	}
}
