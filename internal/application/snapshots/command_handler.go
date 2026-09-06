package snapshots

import "context"

// CommandHandler dispatches sanitized ingestion through the publication service.
type CommandHandler struct { // Service supplies the application behavior invoked by this handler.
	Service *Service
}

// Handle dispatches its command or query through the injected service; ctx controls cancellation.
func (h CommandHandler) Handle(ctx context.Context, c IngestCommand) (IngestResult, error) {
	return h.Service.Ingest(ctx, c)
}
