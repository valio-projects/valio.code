// Package surreal is the sole persistent storage adapter. Each client is bound
// to a database; requests never mutate a shared connection's workspace context.
package surreal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var identifier = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,63}$`)
var ErrNotFound = errors.New("record not found")

type Config struct {
	Endpoint     string
	Namespace    string
	Database     string
	Username     string
	Password     string
	DatabaseAuth bool
}
type Client struct {
	config Config
	http   *http.Client
}
type Statement struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
}

func New(c Config) (*Client, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid SurrealDB endpoint")
	}
	if !identifier.MatchString(c.Namespace) || !identifier.MatchString(c.Database) {
		return nil, errors.New("invalid database scope")
	}
	c.Endpoint = u.Scheme + "://" + u.Host
	return &Client{config: c, http: &http.Client{Timeout: 30 * time.Second}}, nil
}
func (c *Client) Database() string { return c.config.Database }
func (c *Client) ForDatabase(database string) (*Client, error) {
	cfg := c.config
	cfg.Database = database
	return New(cfg)
}

// Query uses the JSON RPC query method. Values travel as bound parameters,
// never interpolated SQL or URL parameters (source payloads can be large).
func (c *Client) Query(ctx context.Context, sql string, vars map[string]any) (statements []Statement, err error) {
	ctx, span := otel.Tracer("valio.infrastructure.surreal").Start(ctx, "surreal.query")
	span.SetAttributes(attribute.String("db.system.name", "surrealdb"), attribute.String("db.namespace", c.config.Database))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "database_failure")
		}
		span.End()
	}()
	if vars == nil {
		vars = map[string]any{}
	}
	body, err := json.Marshal(map[string]any{"id": "query", "method": "query", "params": []any{sql, vars}})
	if err != nil {
		return nil, errors.New("cannot encode database parameters")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint+"/rpc", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Surreal-NS", c.config.Namespace)
	req.Header.Set("Surreal-DB", c.config.Database)
	if c.config.DatabaseAuth {
		req.Header.Set("Surreal-Auth-NS", c.config.Namespace)
		req.Header.Set("Surreal-Auth-DB", c.config.Database)
	}
	req.SetBasicAuth(c.config.Username, c.config.Password)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("database transport: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("database HTTP status %d", res.StatusCode)
	}
	// Bound responses to prevent an accidental unbounded graph/source load.
	data, err := io.ReadAll(io.LimitReader(res.Body, 32*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 32*1024*1024 {
		return nil, errors.New("database response exceeds 32 MiB budget")
	}
	var rpc struct {
		Result []Statement `json:"result"`
		Error  *struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err = json.Unmarshal(data, &rpc); err != nil {
		return nil, errors.New("invalid database RPC response")
	}
	if rpc.Error != nil {
		return nil, fmt.Errorf("database RPC error %d", rpc.Error.Code)
	}
	for i, s := range rpc.Result {
		if s.Status != "OK" {
			return nil, fmt.Errorf("database statement %d failed", i)
		}
	}
	return rpc.Result, nil
}
func DecodeLast[T any](rows []Statement) (T, error) {
	var out T
	if len(rows) == 0 {
		return out, errors.New("empty database response")
	}
	err := json.Unmarshal(rows[len(rows)-1].Result, &out)
	return out, err
}
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Query(ctx, "RETURN true;", nil)
	return err
}
