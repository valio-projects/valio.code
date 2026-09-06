package surreal

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSurrealPersistenceAndScope(t *testing.T) {
	endpoint := os.Getenv("VALIO_TEST_DB_URL")
	if endpoint == "" {
		t.Skip("requires Docker Compose SurrealDB")
	}
	ctx := context.Background()
	name := fmt.Sprintf("test_store_%d", time.Now().UnixNano())
	c, err := New(Config{Endpoint: endpoint, Namespace: "valio_test", Database: name, Username: "root", Password: os.Getenv("VALIO_TEST_DB_PASSWORD")})
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = c.Query(ctx, "REMOVE DATABASE "+name+";", nil) })
	type payload struct {
		Name string `json:"name"`
	}
	if err = c.Create(ctx, "source_snapshot", "first", payload{"snapshot"}); err != nil {
		t.Fatal(err)
	}
	if err = c.Create(ctx, "source_snapshot", "first", payload{"replacement"}); err == nil {
		t.Fatal("immutable record replaced")
	}
	got, err := Get[payload](ctx, c, "source_snapshot", "first")
	if err != nil || got.Name != "snapshot" {
		t.Fatalf("stored: %+v %v", got, err)
	}
	if err = c.ProvisionUser(ctx, "api_test", "integration-only-long-password", "EDITOR"); err != nil {
		t.Fatal(err)
	}
	cfg := c.config
	cfg.Username = "api_test"
	cfg.Password = "integration-only-long-password"
	cfg.DatabaseAuth = true
	api, _ := New(cfg)
	if err = api.Ping(ctx); err != nil {
		t.Fatal("database-level authentication:", err)
	}
	other, _ := api.ForDatabase(name + "_other")
	if err = other.Ping(ctx); err == nil {
		t.Fatal("database account crossed workspace boundary")
	}
}
