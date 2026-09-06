package surreal

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
)

//go:embed migrations/*.surql
var migrations embed.FS

// Migrate is deliberately separate from API startup. Use provisioning credentials.
func (c *Client) Migrate(ctx context.Context) error {
	_, err := c.Query(ctx, "DEFINE NAMESPACE IF NOT EXISTS "+c.config.Namespace+"; USE NS "+c.config.Namespace+"; DEFINE DATABASE IF NOT EXISTS "+c.config.Database+";", nil)
	if err != nil {
		return err
	}
	data, err := migrations.ReadFile("migrations/001_initial.surql")
	if err != nil {
		return err
	}
	_, err = c.Query(ctx, string(data), nil)
	return err
}

// ProvisionUser creates a database-scoped system user. Role is intentionally
// constrained rather than accepting arbitrary SurrealQL from configuration.
func (c *Client) ProvisionUser(ctx context.Context, name, password, role string) error {
	if !identifier.MatchString(name) || len(password) < 16 || (role != "EDITOR" && role != "VIEWER") {
		return errors.New("invalid database user configuration")
	}
	// SurrealDB 3.2 DEFINE USER requires a string literal, not a parameter.
	// JSON quoting escapes every delimiter; Query never logs SQL or DB messages.
	quoted, _ := json.Marshal(password)
	_, err := c.Query(ctx, "DEFINE USER OVERWRITE "+name+" ON DATABASE PASSWORD "+string(quoted)+" ROLES "+role+";", nil)
	return err
}
