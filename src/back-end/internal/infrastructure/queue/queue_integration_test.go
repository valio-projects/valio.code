package queue

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
)

func TestSurrealLeaseFencingAndIdempotency(t *testing.T) {
	endpoint := os.Getenv("VALIO_TEST_DB_URL")
	if endpoint == "" {
		t.Skip("requires Docker Compose SurrealDB; scripts/test-integration.ps1")
	}
	ctx := context.Background()
	name := fmt.Sprintf("test_jobs_%d", time.Now().UnixNano())
	db, err := surreal.New(surreal.Config{Endpoint: endpoint, Namespace: "valio_test", Database: name, Username: "root", Password: os.Getenv("VALIO_TEST_DB_PASSWORD")})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Query(ctx, "REMOVE DATABASE "+name+";", nil) })
	q := Queue{DB: db}
	id, err := q.Enqueue(ctx, "analysis", "same-source", map[string]string{"snapshot": "one"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := q.Enqueue(ctx, "analysis", "same-source", map[string]string{"snapshot": "one"}, 1)
	if err != nil || repeated != id {
		t.Fatalf("idempotency: %s %v", repeated, err)
	}
	first, err := q.Claim(ctx, "worker-one")
	if err != nil || first == nil {
		debug, _ := db.Query(ctx, "SELECT * FROM job;", nil)
		t.Logf("fixture jobs: %s", debug[len(debug)-1].Result)
		t.Fatalf("claim: %+v %v", first, err)
	}
	blocked, err := q.Claim(ctx, "worker-two")
	if err != nil || blocked != nil {
		t.Fatalf("duplicate ownership: %+v %v", blocked, err)
	}
	if err = q.Heartbeat(ctx, *first); err != nil {
		t.Fatal(err)
	}
	_, err = db.Query(ctx, "UPDATE type::record('job',$key) SET lease_until=0;", map[string]any{"key": id})
	if err != nil {
		t.Fatal(err)
	}
	second, err := q.Claim(ctx, "worker-two")
	if err != nil || second == nil {
		t.Fatalf("reclaim: %+v %v", second, err)
	}
	if second.Fence <= first.Fence {
		t.Fatal("fence did not advance")
	}
	if err = q.Complete(ctx, *first); err != ErrLeaseLost {
		t.Fatalf("stale worker completed: %v", err)
	}
	if err = q.Publish(ctx, *first, "text", "stale"); err == nil {
		t.Fatal("stale generation published")
	}
	if err = q.Publish(ctx, *second, "text", "new"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(ctx, "SELECT VALUE generation FROM projection_head;", nil)
	if err != nil {
		t.Fatal(err)
	}
	generations, err := surreal.DecodeLast[[]string](rows)
	if err != nil || len(generations) != 1 || generations[0] != "new" {
		t.Fatalf("publication: %v %v", generations, err)
	}
}
