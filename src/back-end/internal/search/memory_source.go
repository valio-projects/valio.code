package search

import "context"

// MemorySource is an in-memory stable Source useful for tests and small queries.
type MemorySource []File

// Files returns a copy of its file slice unless ctx is already cancelled.
func (m MemorySource) Files(ctx context.Context) ([]File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []File(m), nil
}
