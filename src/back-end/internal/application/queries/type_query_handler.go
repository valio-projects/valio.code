package queries

import "context"

// TypeHandler is the transport-facing contract for scoped type lookups.
type TypeHandler interface {
	// Handle dispatches its command or query through the injected service; ctx controls cancellation.
	Handle(context.Context, TypeQuery) (TypeResult, error)
}

// TypeQueryHandler dispatches a type lookup through the immutable query service.
type TypeQueryHandler struct { // Service supplies the application behavior invoked by this handler.
	Service *Service
}

// Handle dispatches its command or query through the injected service; ctx controls cancellation.
func (h TypeQueryHandler) Handle(ctx context.Context, q TypeQuery) (TypeResult, error) {
	return h.Service.Types(ctx, q)
}
