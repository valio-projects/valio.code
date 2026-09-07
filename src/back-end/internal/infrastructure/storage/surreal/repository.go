package surreal

import (
	"context"
	"errors"
)

// Repository is a typed adapter over one fixed table and database. Application
// services depend on their own narrow repository interfaces, not on Query.
type Repository[T any] struct {
	client    *Client
	table     string
	immutable bool
}

func NewRepository[T any](client *Client, table string, immutable bool) (*Repository[T], error) {
	if client == nil || !identifier.MatchString(table) {
		return nil, errors.New("invalid repository configuration")
	}
	return &Repository[T]{client: client, table: table, immutable: immutable}, nil
}
func (r *Repository[T]) Find(ctx context.Context, key string) (T, error) {
	return Get[T](ctx, r.client, r.table, key)
}
func (r *Repository[T]) List(ctx context.Context, limit, offset int) ([]T, error) {
	return List[T](ctx, r.client, r.table, limit, offset)
}
func (r *Repository[T]) Create(ctx context.Context, key string, value T) error {
	return r.client.Create(ctx, r.table, key, value)
}
func (r *Repository[T]) Save(ctx context.Context, key string, value T) error {
	if r.immutable {
		return errors.New("immutable repository requires Create; replacements are forbidden")
	}
	return r.client.Put(ctx, r.table, key, value)
}
