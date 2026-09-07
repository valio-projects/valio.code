package surreal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/application/queries"
	app "github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/projects"
	"os"
	"testing"
	"time"
)

func safeTestSnapshot(content string) agent.Snapshot {
	sum := sha256.Sum256([]byte(content))
	s := agent.Snapshot{Version: 1, Files: []agent.File{{Path: "main.go", Hash: hex.EncodeToString(sum[:]), Size: len(content), Language: "go", Content: content}}}
	b, _ := json.Marshal(s)
	sum = sha256.Sum256(b)
	s.ID = hex.EncodeToString(sum[:])
	return s
}
func TestAppPublicationPinnedMembershipAndTypes(t *testing.T) {
	endpoint := os.Getenv("VALIO_TEST_DB_URL")
	if endpoint == "" {
		t.Skip("requires real SurrealDB")
	}
	ctx := context.Background()
	name := fmt.Sprintf("test_app_%d", time.Now().UnixNano())
	client, e := New(Config{Endpoint: endpoint, Namespace: "valio_test", Database: name, Username: "root", Password: os.Getenv("VALIO_TEST_DB_PASSWORD")})
	if e != nil {
		t.Fatal(e)
	}
	if e = client.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _, _ = client.Query(ctx, "REMOVE DATABASE "+name+";", nil) })
	store, e := NewAppStore(client, "workspace-main")
	if e != nil {
		t.Fatal(e)
	}
	workspace := domain.Workspace{ID: "workspace-main", Name: "test"}
	if e = store.Bootstrap(ctx, workspace); e != nil {
		t.Fatal(e)
	}
	admin := catalog.Service{Store: store, Workspace: workspace}
	_, e = admin.Register(ctx, domain.Repository{ID: "repo", WorkspaceID: workspace.ID})
	if e != nil {
		t.Fatal(e)
	}
	def := projects.Definition{Project: domain.Project{ID: "p", WorkspaceID: workspace.ID, Key: "p", Name: "Project", Kind: domain.ProjectLibrary, Status: domain.ProjectActive}, Roots: []domain.ProjectSourceRoot{{RepositoryID: "repo", Path: ".", Version: "v1"}}}
	if _, e = admin.SaveProject(ctx, def, false); e != nil {
		t.Fatal(e)
	}
	ingester := app.Service{Store: store, WorkspaceID: workspace.ID}
	query := queries.Service{Store: store, WorkspaceID: workspace.ID}
	command := app.IngestCommand{WorkspaceID: workspace.ID, RepositoryID: "repo", Snapshot: safeTestSnapshot("package main\ntype Item struct { Name string }\n")}
	receipt, e := ingester.Ingest(ctx, command)
	if e != nil {
		t.Fatal("ingest", e)
	}
	if receipt.Status != "partial" {
		t.Fatal(receipt)
	}
	again, e := ingester.Ingest(ctx, command)
	if e != nil || again != receipt {
		t.Fatalf("duplicate: %+v %v", again, e)
	}
	types, e := query.Types(ctx, queries.TypeQuery{Name: "Item", ViewID: receipt.ViewID})
	if e != nil || len(types.Candidates) != 1 {
		t.Fatalf("types: %+v %v", types, e)
	}
	result, e := query.Search(ctx, queries.SearchQuery{Query: "Item", Scope: queries.SearchScope{WorkspaceID: workspace.ID, ViewID: receipt.ViewID, ProjectIDs: []string{"p"}}})
	if e != nil || result.Total != 1 {
		t.Fatalf("search: %+v %v", result, e)
	}
	def.Project.Name = "Renamed"
	def.Roots[0].Path = "other"
	if _, e = admin.SaveProject(ctx, def, true); e != nil {
		t.Fatal(e)
	}
	next, e := ingester.Ingest(ctx, command)
	if e != nil || next.ViewID == receipt.ViewID {
		t.Fatalf("definition creates view: %+v %v", next, e)
	}
	old, e := query.Search(ctx, queries.SearchQuery{Query: "Item", Scope: queries.SearchScope{WorkspaceID: workspace.ID, ViewID: receipt.ViewID, ProjectIDs: []string{"p"}}})
	if e != nil || old.Total != 1 {
		t.Fatalf("old membership changed: %+v %v", old, e)
	}
	fresh, e := query.Search(ctx, queries.SearchQuery{Query: "Item", Scope: queries.SearchScope{WorkspaceID: workspace.ID, ViewID: next.ViewID, ProjectIDs: []string{"p"}}})
	if e != nil || fresh.Total != 0 {
		t.Fatalf("new membership: %+v %v", fresh, e)
	}
	oldView, e := store.View(ctx, receipt.ViewID)
	if e != nil || oldView.Projects[0].Project.Name != "Project" {
		t.Fatal("pinned project changed", e)
	}
	// Staged records roll back if any immutable entry conflicts; no partial view
	// or newly staged blob can leak through the publication pointer.
	current, e := store.Latest(ctx)
	if e != nil {
		t.Fatal(e)
	}
	broken := current
	broken.ID = "unpublished"
	broken.Files = append([]snapshots.FileRef{}, current.Files...)
	broken.Files[0].BlobID = "wrong-blob"
	e = store.Publish(ctx, snapshots.Publication{ExpectedHead: current.ID, Snapshot: snapshots.Metadata{ID: "staged-snapshot", WorkspaceID: workspace.ID}, View: broken, Blobs: []snapshots.Blob{{ID: "staged-blob", WorkspaceID: workspace.ID, Content: "staged"}}})
	if !errors.Is(e, fault.ErrConflict) {
		t.Fatal("expected publication conflict", e)
	}
	if _, e = Get[snapshots.Blob](ctx, client, "source_blob", "staged-blob"); !errors.Is(e, ErrNotFound) {
		t.Fatal("partial staged blob visible", e)
	}
	if _, e = store.View(ctx, "unpublished"); !errors.Is(e, fault.ErrNotFound) {
		t.Fatal("partial view visible", e)
	}
	latest, e := store.Latest(ctx)
	if e != nil || latest.ID != current.ID {
		t.Fatal("failed transaction changed head", e)
	}
	if _, e = admin.Register(ctx, domain.Repository{ID: "repo-two", WorkspaceID: workspace.ID}); e != nil {
		t.Fatal(e)
	}
	def.Roots = []domain.ProjectSourceRoot{{RepositoryID: "repo", Path: ".", Version: "v2"}, {RepositoryID: "repo-two", Path: ".", Version: "v1"}}
	if _, e = admin.SaveProject(ctx, def, true); e != nil {
		t.Fatal(e)
	}
	two, e := ingester.Ingest(ctx, app.IngestCommand{WorkspaceID: workspace.ID, RepositoryID: "repo-two", Snapshot: safeTestSnapshot("package second\ntype Second struct { Count int }\n")})
	if e != nil {
		t.Fatal("multi repository ingestion", e)
	}
	mixed, e := store.View(ctx, two.ViewID)
	if e != nil || !mixed.Mixed || len(mixed.Repositories) != 2 {
		t.Fatalf("mixed view %+v %v", mixed, e)
	}
	both, e := query.Search(ctx, queries.SearchQuery{Query: "type", Scope: queries.SearchScope{WorkspaceID: workspace.ID, ViewID: two.ViewID, ProjectIDs: []string{"p"}}})
	if e != nil || both.Total != 2 {
		t.Fatalf("mixed search %+v %v", both, e)
	}
}
