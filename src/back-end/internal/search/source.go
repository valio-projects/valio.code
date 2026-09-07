package search

import "context"

// Source must return a stable snapshot without a hidden candidate limit.
type Source interface {
	Files(context.Context) ([]File, error)
}
