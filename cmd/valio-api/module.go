package main

import (
	"context"
	"errors"
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/authorization"
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"github.com/valio-projects/valio.code/internal/infrastructure/telemetry"
	httpapi "github.com/valio-projects/valio.code/internal/transport/http"
	mcpapi "github.com/valio-projects/valio.code/internal/transport/mcp"
	"go.uber.org/fx"
	"net"
	"net/http"
	"time"
)

func apiModule() fx.Option {
	return fx.Module("api", fx.Provide(apiConfig, configuration.Database, surreal.New, newStore, newAuthorization, newCatalog, newQueries, newIngestion, newMCP, newHTTP), fx.Invoke(serve))
}
func newStore(db *surreal.Client, c apiConfiguration) (*surreal.AppStore, error) {
	return surreal.NewAppStore(db, c.Workspace.ID)
}
func newAuthorization(c apiConfiguration) (*authorization.Bootstrap, error) {
	return authorization.NewBootstrap(c.Token, c.Origins)
}
func newCatalog(store *surreal.AppStore, c apiConfiguration) *catalog.Service {
	return &catalog.Service{Store: store, Workspace: c.Workspace}
}
func newQueries(store *surreal.AppStore, c apiConfiguration) *queries.Service {
	return &queries.Service{Store: store, WorkspaceID: c.Workspace.ID}
}
func newIngestion(store *surreal.AppStore, c apiConfiguration) *snapshots.Service {
	return &snapshots.Service{Store: store, WorkspaceID: c.Workspace.ID}
}
func newMCP(store *surreal.AppStore, q *queries.Service) (*mcpapi.Server, error) {
	return mcpapi.NewServer(store, *q)
}
func newHTTP(a *authorization.Bootstrap, c *catalog.Service, q *queries.Service, i *snapshots.Service, db *surreal.Client, mcp *mcpapi.Server) *httpapi.Server {
	return &httpapi.Server{Auth: a, Catalog: c, Queries: q, Ingestion: i, Ready: db.Ping, MCP: mcp.HTTPHandler()}
}
func serve(lifecycle fx.Lifecycle, shutdown fx.Shutdowner, c apiConfiguration, api *httpapi.Server, store *surreal.AppStore) {
	server := &http.Server{Addr: c.Address, Handler: telemetry.HTTPMiddleware(api.Handler()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		if e := store.Bootstrap(ctx, c.Workspace); e != nil {
			return e
		}
		listener, e := net.Listen("tcp", c.Address)
		if e != nil {
			return e
		}
		go func() {
			if e := server.Serve(listener); e != nil && !errors.Is(e, http.ErrServerClosed) {
				_ = shutdown.Shutdown(fx.ExitCode(1))
			}
		}()
		return nil
	}, OnStop: func(ctx context.Context) error { return server.Shutdown(ctx) }})
}
