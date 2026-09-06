package surreal

import (
	"context"
	"errors"
)

func (c *Client) Put(ctx context.Context, table, key string, payload any) error {
	if !identifier.MatchString(table) || key == "" {
		return errors.New("invalid record identity")
	}
	_, err := c.Query(ctx, "UPSERT type::record($table,$key) SET key=$key, payload=$payload;", map[string]any{"table": table, "key": key, "payload": payload})
	return err
}
func (c *Client) Create(ctx context.Context, table, key string, payload any) error {
	if !identifier.MatchString(table) || key == "" {
		return errors.New("invalid record identity")
	}
	_, err := c.Query(ctx, "CREATE type::record($table,$key) SET key=$key, payload=$payload;", map[string]any{"table": table, "key": key, "payload": payload})
	return err
}
func Get[T any](ctx context.Context, c *Client, table, key string) (T, error) {
	var out T
	if !identifier.MatchString(table) {
		return out, errors.New("invalid table")
	}
	rows, err := c.Query(ctx, "SELECT VALUE payload FROM type::record($table,$key);", map[string]any{"table": table, "key": key})
	if err != nil {
		return out, err
	}
	vals, err := DecodeLast[[]T](rows)
	if err != nil {
		return out, err
	}
	if len(vals) == 0 {
		return out, ErrNotFound
	}
	return vals[0], nil
}
func List[T any](ctx context.Context, c *Client, table string, limit, offset int) ([]T, error) {
	if !identifier.MatchString(table) || limit < 1 || limit > 1000 || offset < 0 {
		return nil, errors.New("invalid list request")
	}
	rows, err := c.Query(ctx, "SELECT key,payload FROM type::table($table) ORDER BY key LIMIT $limit START $offset;", map[string]any{"table": table, "limit": limit, "offset": offset})
	if err != nil {
		return nil, err
	}
	vals, err := DecodeLast[[]struct {
		Payload T `json:"payload"`
	}](rows)
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, len(vals))
	for _, v := range vals {
		out = append(out, v.Payload)
	}
	return out, nil
}
