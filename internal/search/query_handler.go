package search

import "context"

// QueryHandler applies request-local scope and budgets to a configured engine.
// It is read-only and never mutates the shared source or engine configuration.
type QueryHandler struct{ Engine Engine }

func (handler QueryHandler) Handle(ctx context.Context, query SearchQuery) (Result, error) {
	engine := handler.Engine
	engine.Options = query.Options
	return engine.Search(ctx, query.Expression)
}
