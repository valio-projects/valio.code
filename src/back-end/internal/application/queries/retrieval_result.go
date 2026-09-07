package queries

import "github.com/valio-projects/valio.code/internal/retrieval"

// RetrievalResult retains channel provenance and explicit missing capabilities.
type RetrievalResult struct {
	ViewID         string                  `json:"viewId"`
	Mode           string                  `json:"mode"`
	Hits           []retrieval.ScoredChunk `json:"hits"`
	CandidateCount int                     `json:"candidateCount"`
	Truncated      bool                    `json:"truncated"`
	Diagnostics    []string                `json:"diagnostics"`
}
