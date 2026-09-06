package search

import "context"

type MemorySource []File

func (m MemorySource) Files(ctx context.Context) ([]File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []File(m), nil
}
