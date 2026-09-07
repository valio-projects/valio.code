// Package scheduling implements durable at-least-once jobs in SurrealDB.
package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"go.opentelemetry.io/otel/propagation"
)

type Queue struct{ DB *surreal.Client }

func (q Queue) Enqueue(ctx context.Context, kind, key string, payload any, priority int) (string, error) {
	if kind == "" || key == "" {
		return "", errors.New("kind and idempotency key required")
	}
	sum := sha256.Sum256([]byte(kind + "\x00" + key))
	id := hex.EncodeToString(sum[:])
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)
	// Ignore duplicates without resetting a running/completed job.
	_, err := q.DB.Query(ctx, `INSERT IGNORE INTO job {
 id:type::record('job',$key),key:$key,kind:$kind,payload:$payload,
 status:'queued',priority:$priority,attempts:0,fence:0,owner:'',lease_until:0,
 available_at:time::unix(time::now()),error_code:'',trace_parent:$trace_parent
 };`, map[string]any{"key": id, "kind": kind, "payload": payload, "priority": priority, "trace_parent": carrier.Get("traceparent")})
	return id, err
}

// Claim selects and updates in one transaction. Optimistic
// conflicts can be retried by the worker; no process-local lock is authoritative.
func (q Queue) Claim(ctx context.Context, owner string) (*Job, error) {
	if owner == "" {
		return nil, errors.New("worker owner required")
	}
	rows, err := q.DB.Query(ctx, `BEGIN TRANSACTION;
 LET $now=time::unix(time::now());
 LET $candidate=(SELECT * FROM job WITH NOINDEX WHERE
 (status='queued' AND available_at <= $now) OR (status='running' AND lease_until <= $now)
 ORDER BY priority DESC, available_at ASC LIMIT 1);
 LET $claimed=(UPDATE $candidate.id SET status='running',owner=$owner,
 lease_until=$now+60,fence=fence+1,attempts=attempts+1
 RETURN AFTER);
 RETURN $claimed;
 COMMIT TRANSACTION;`, map[string]any{"owner": owner})
	if err != nil {
		return nil, err
	}
	// Transaction control statements may also be returned by the transport.
	for i := len(rows) - 1; i >= 0; i-- {
		var jobs []Job
		if json.Unmarshal(rows[i].Result, &jobs) == nil && len(jobs) > 0 && jobs[0].ID != "" {
			return &jobs[0], nil
		}
	}
	return nil, nil
}
