package surreal

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	appembeddings "github.com/valio-projects/valio.code/internal/application/embeddings"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

func TestEmbeddingRepositoryIntegration(t *testing.T) {
	client := embeddingIntegrationClient(t)
	ctx := context.Background()
	repository, err := NewEmbeddingRepository(client)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := embeddings.NewProfile("code-v1", embeddings.Code, 1, "test", "model", "revision", "cosine", 2)
	if err != nil {
		t.Fatal(err)
	}
	otherProfile, err := embeddings.NewProfile("code-v2", embeddings.Code, 1, "test", "other-model", "revision", "cosine", 2)
	if err != nil {
		t.Fatal(err)
	}
	input := embeddingInput("selected", "workspace", "view", "approved selected")
	record, err := embeddings.NewRecord(profile, input, []float32{1, 2}, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Save(ctx, record); err != nil {
		t.Fatal(err)
	}
	// Unrelated keys are deliberately inserted before the scoped read. The single
	// matching row must still be returned with limit one, proving filters precede limit.
	for _, other := range []embeddings.Input{
		embeddingInput("other-workspace", "other-workspace", "view", "approved other workspace"),
		embeddingInput("other-view", "workspace", "other-view", "approved other view"),
	} {
		otherRecord, createErr := embeddings.NewRecord(profile, other, []float32{1, 2}, time.Unix(1, 0))
		if createErr != nil {
			t.Fatal("create other scoped record", createErr)
		}
		if saveErr := repository.Save(ctx, otherRecord); saveErr != nil {
			t.Fatal("save other scoped record", saveErr)
		}
	}
	otherProfileRecord, err := embeddings.NewRecord(otherProfile, input, []float32{1, 2}, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Save(ctx, otherProfileRecord); err != nil {
		t.Fatal(err)
	}
	if _, err = appembeddings.RankCosine(profile, []float32{1, 2}, []embeddings.Record{otherProfileRecord}, 1); err == nil {
		t.Fatal("ranking accepted a candidate from a different profile")
	}
	listed, err := repository.List(ctx, "workspace", "view", profile.Fingerprint, 1)
	if err != nil || len(listed) != 1 || listed[0].ID != record.ID {
		t.Fatalf("filtered list: %+v %v", listed, err)
	}
	if _, err = repository.List(ctx, "workspace", "view", otherProfile.Fingerprint, 1); err != nil {
		t.Fatal("second profile must have an independently scoped list", err)
	}
	changed := record
	changed.Vector = []float32{2, 1}
	if err = repository.Save(ctx, changed); err == nil {
		t.Fatal("changed vector replaced immutable record")
	}
	if err = repository.Save(ctx, record); err != nil {
		t.Fatal("identical immutable record was not idempotent", err)
	}

	concurrentInput := embeddingInput("concurrent", "workspace", "view", "approved concurrent")
	concurrentRecord, err := embeddings.NewRecord(profile, concurrentInput, []float32{1, 2}, time.Unix(2, 0))
	if err != nil {
		t.Fatal(err)
	}
	errors := make(chan error, 8)
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			errors <- repository.Save(ctx, concurrentRecord)
		}()
	}
	group.Wait()
	close(errors)
	for saveErr := range errors {
		if saveErr != nil {
			t.Fatal("concurrent idempotent save", saveErr)
		}
	}

	embedder := &countingEmbedder{}
	service := appembeddings.Service{Repository: repository, Provider: embedder, Now: func() time.Time { return time.Unix(3, 0) }}
	computedInput := embeddingInput("computed", "workspace", "view", "approved compute")
	first, err := service.Compute(ctx, profile, computedInput)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Compute(ctx, profile, computedInput)
	if err != nil || second.ID != first.ID || len(second.Vector) != 2 || embedder.calls != 1 {
		t.Fatalf("cached compute: %+v %v calls=%d", second, err, embedder.calls)
	}
}

func embeddingIntegrationClient(t *testing.T) *Client {
	t.Helper()
	endpoint := os.Getenv("VALIO_TEST_DB_URL")
	if endpoint == "" {
		t.Skip("requires Docker Compose SurrealDB through VALIO_TEST_DB_URL")
	}
	ctx := context.Background()
	database := fmt.Sprintf("test_embeddings_%d", time.Now().UnixNano())
	client, err := New(Config{Endpoint: endpoint, Namespace: "valio_test", Database: database, Username: "root", Password: os.Getenv("VALIO_TEST_DB_PASSWORD")})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = client.Query(ctx, "REMOVE DATABASE "+database+";", nil) })
	return client
}

func embeddingInput(id string, workspace domain.WorkspaceID, view domain.ViewID, representation string) embeddings.Input {
	return embeddings.Input{ID: id, WorkspaceID: workspace, ViewID: view, Kind: embeddings.Code, Fingerprint: embeddings.Hash(representation), Representation: representation, PolicyApproved: true}
}

type countingEmbedder struct{ calls int }

func (e *countingEmbedder) Embed(context.Context, embeddings.Profile, string) ([]float32, error) {
	e.calls++
	return []float32{1, 2}, nil
}
