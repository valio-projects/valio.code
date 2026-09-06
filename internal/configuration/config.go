package configuration

import (
	"errors"
	"os"

	"github.com/valio-projects/valio.code/internal/storage/surreal"
)

func Env(key, fallback string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return fallback
}
func Database() (surreal.Config, error) {
	c := surreal.Config{Endpoint: Env("VALIO_DB_URL", "http://127.0.0.1:18000"), Namespace: Env("VALIO_DB_NAMESPACE", "valio"), Database: Env("VALIO_DB_DATABASE", "workspace_main"), Username: Env("VALIO_DB_USER", "valio_api"), Password: os.Getenv("VALIO_DB_PASSWORD"), DatabaseAuth: Env("VALIO_DB_AUTH_LEVEL", "database") == "database"}
	if c.Password == "" {
		return c, errors.New("VALIO_DB_PASSWORD is required")
	}
	return c, nil
}
