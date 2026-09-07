package configuration

import (
	"os"

	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
)

const (
	// DefaultDatabaseURL is the loopback SurrealDB endpoint used by the local stack.
	DefaultDatabaseURL = "http://127.0.0.1:18000"
	// DefaultDatabaseNamespace scopes the local Valio data in SurrealDB.
	DefaultDatabaseNamespace = "valio"
	// DefaultDatabaseName identifies the local Valio workspace database.
	DefaultDatabaseName = "workspace_main"
	// DefaultDatabaseUser is the SurrealDB root user used by the local stack.
	DefaultDatabaseUser = "root"
	// DefaultDatabasePassword is the development-only root credential used by the local stack.
	DefaultDatabasePassword = "valio"
	// DefaultWorkspaceID identifies the workspace created by the local API.
	DefaultWorkspaceID = "workspace-main"
	// DefaultAPIToken is the development-only credential shared by local clients and the API.
	DefaultAPIToken = "valio-local-development-token-0001"
	// DefaultAPIAddress is the loopback address used by the local API.
	DefaultAPIAddress = "127.0.0.1:8090"
	// DefaultServerURL addresses the Compose ingress used by native clients.
	DefaultServerURL = "http://127.0.0.1:8080"
)

// Env returns the non-empty value of key, or fallback when key is unset or empty.
func Env(key, fallback string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return fallback
}

// ApplicationToken returns the API credential for local clients and the API.
// VALIO_API_TOKEN overrides the development default.
func ApplicationToken() string { return Env("VALIO_API_TOKEN", DefaultAPIToken) }

// Database returns the root-authenticated SurrealDB configuration for the local
// stack. VALIO_DB_* values override each field. Integration tests construct
// their isolated configurations directly and never alter runtime defaults.
func Database() (surreal.Config, error) {
	return surreal.Config{
		Endpoint:     Env("VALIO_DB_URL", DefaultDatabaseURL),
		Namespace:    Env("VALIO_DB_NAMESPACE", DefaultDatabaseNamespace),
		Database:     Env("VALIO_DB_DATABASE", DefaultDatabaseName),
		Username:     Env("VALIO_DB_USER", DefaultDatabaseUser),
		Password:     Env("VALIO_DB_PASSWORD", DefaultDatabasePassword),
		DatabaseAuth: Env("VALIO_DB_AUTH_LEVEL", "root") == "database",
	}, nil
}
