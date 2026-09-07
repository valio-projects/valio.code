package queries

import "context"

// SearchHandler defines transport-independent bounded search dispatch.
type SearchHandler interface {
	// Handle dispatches its command or query through the injected service; ctx controls cancellation.
	Handle(context.Context, SearchQuery) (SearchResult, error)
}

// SearchQueryHandler dispatches search using pinned application queries.
type SearchQueryHandler struct { // Service supplies the application behavior invoked by this handler.
	Service *Service
}

// Handle dispatches its command or query through the injected service; ctx controls cancellation.
func (h SearchQueryHandler) Handle(ctx context.Context, q SearchQuery) (SearchResult, error) {
	return h.Service.Search(ctx, q)
}
