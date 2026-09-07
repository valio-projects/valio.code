package surreal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/application/catalog"
	"github.com/valio-projects/valio.code/internal/application/queries"
	appsnapshots "github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	syntaxadapter "github.com/valio-projects/valio.code/internal/infrastructure/syntax"
	"github.com/valio-projects/valio.code/internal/projects"
	"os"
	"testing"
)

type integrationModels struct{}

func (integrationModels) ResolveModel(name string, kind embeddings.Kind) (embeddings.Profile, embeddings.Embedder, error) {
	p, e := embeddings.NewProfile("test", kind, 1, "test", "test-model", "v1", "cosine", 3)
	return p, integrationModels{}, e
}
func (integrationModels) Embed(context.Context, embeddings.Profile, string) ([]float32, error) {
	return []float32{1, 2, 3}, nil
}

func TestIntelligencePublicationThroughSurreal(t *testing.T) {
	client := embeddingIntegrationClient(t)
	ctx := context.Background()
	store, err := NewAppStore(client, "workspace-main")
	if err != nil {
		t.Fatal(err)
	}
	workspace := domain.Workspace{ID: "workspace-main", Name: "integration"}
	if err = store.Bootstrap(ctx, workspace); err != nil {
		t.Fatal(err)
	}
	admin := catalog.Service{Store: store, Workspace: workspace}
	if _, err = admin.Register(ctx, domain.Repository{ID: "repo", WorkspaceID: workspace.ID}); err != nil {
		t.Fatal(err)
	}
	def := projects.Definition{Project: domain.Project{ID: "billing", WorkspaceID: workspace.ID, Key: "billing", Name: "Billing", Kind: domain.ProjectService, Status: domain.ProjectActive}, Roots: []domain.ProjectSourceRoot{{RepositoryID: "repo", Path: ".", Version: "v1"}}}
	if _, err = admin.SaveProject(ctx, def, false); err != nil {
		t.Fatal(err)
	}
	source := agent.Snapshot{Version: 1, Files: []agent.File{}}
	for _, file := range []struct{ path, language, text string }{{"a.go", "go", "package billing\ntype Order struct { ID int }\nfunc Charge(o Order) int { return Save(o.ID) }\n"}, {"b.go", "go", "package billing\nfunc Save(id int) int { return id + 1 }\n"}, {"README.md", "markdown", "Billing stores orders and charges customers.\n"}} {
		sum := sha256.Sum256([]byte(file.text))
		source.Files = append(source.Files, agent.File{Path: file.path, Language: file.language, Content: file.text, Size: len(file.text), Hash: hex.EncodeToString(sum[:])})
	}
	encoded, _ := json.Marshal(source)
	sum := sha256.Sum256(encoded)
	source.ID = hex.EncodeToString(sum[:])
	ingester := appsnapshots.Service{Store: store, WorkspaceID: workspace.ID}
	if script := os.Getenv("VALIO_TEST_SYNTAX_HELPER"); script != "" {
		processor, e := syntaxadapter.New("node", script)
		if e != nil {
			t.Fatal(e)
		}
		ingester.Syntax = processor
		for _, file := range []struct{ path, language, text string }{
			{"user.cs", "csharp", "public class User { public int ID { get; set; } }"},
			{"user.c", "c", "struct User { int id; };"},
			{"user.cpp", "cpp", "class User { public: int id; };"},
			{"user.js", "javascript", "export class User { get() { return 1; } }"},
			{"user.ts", "typescript", "export class User { id: number; }"},
			{"User.java", "java", "public class User { public int id; }"},
		} {
			sum := sha256.Sum256([]byte(file.text))
			source.Files = append(source.Files, agent.File{Path: file.path, Language: file.language, Content: file.text, Size: len(file.text), Hash: hex.EncodeToString(sum[:])})
		}
		source.ID = ""
		b, _ := json.Marshal(source)
		hash := sha256.Sum256(b)
		source.ID = hex.EncodeToString(hash[:])
	}
	command := appsnapshots.IngestCommand{WorkspaceID: workspace.ID, RepositoryID: "repo", Snapshot: source}
	receipt, err := ingester.Ingest(ctx, command)
	if err != nil {
		t.Fatal("publish", err)
	}
	vectors, err := NewEmbeddingRepository(client)
	if err != nil {
		t.Fatal(err)
	}
	service := queries.Service{Store: store, WorkspaceID: workspace.ID, Models: integrationModels{}, Vectors: vectors}
	scope := queries.SearchScope{WorkspaceID: workspace.ID, ViewID: receipt.ViewID, ProjectIDs: []string{"billing"}}
	persistedView, e := store.View(ctx, receipt.ViewID)
	if e != nil {
		t.Fatal(e)
	}
	persistedArtifacts, e := store.Artifacts(ctx, persistedView)
	if e != nil {
		t.Fatal(e)
	}
	for _, artifact := range persistedArtifacts {
		for _, chunk := range artifact.Chunks {
			for _, file := range persistedView.Files {
				if file.ID == chunk.FileID && (chunk.Start < 0 || chunk.End < chunk.Start || chunk.End > file.Size || chunk.Text != "" && len(chunk.Text) != chunk.End-chunk.Start) {
					t.Fatalf("persisted fixture chunk range: file=%s sourceSize=%d range=%d:%d textBytes=%d kind=%s text=%q", file.Path, file.Size, chunk.Start, chunk.End, len(chunk.Text), chunk.Kind, chunk.Text)
				}
			}
		}
	}
	if ingester.Syntax != nil {
		written, e := service.Syntax(ctx, queries.SyntaxQuery{Scope: scope, Name: "User"})
		if e != nil || len(written.Reports) != 6 {
			t.Fatalf("multi-language persisted types: %d %v", len(written.Reports), e)
		}
		found, e := service.Search(ctx, queries.SearchQuery{Scope: scope, Query: "symbol:User"})
		if e != nil || found.Total != 6 {
			t.Fatalf("multi-language symbol metadata: %+v %v", found, e)
		}
		descriptors, e := service.Types(ctx, queries.TypeQuery{Name: "User", ViewID: receipt.ViewID, ProjectID: "billing"})
		if e != nil || len(descriptors.Candidates) != 6 {
			t.Fatalf("multi-language type catalog: %d %v", len(descriptors.Candidates), e)
		}
		outline, e := service.Structure(ctx, queries.StructureQuery{Scope: scope, Name: "User", Depth: 3})
		if e != nil || len(outline.Files) != 6 || outline.Truncated {
			t.Fatalf("multi-language structure: %+v %v", outline, e)
		}
		for _, file := range outline.Files {
			if len(file.Nodes) < 2 || len(file.Edges) == 0 {
				t.Fatalf("missing type members in %s", file.Path)
			}
		}
		chunks, e := service.Retrieve(ctx, queries.RetrievalQuery{Scope: scope, Mode: "symbol", Query: "User"})
		if e != nil || len(chunks.Hits) < 6 {
			t.Fatalf("syntax symbol representations: %d %v", len(chunks.Hits), e)
		}
	}
	graph, err := service.Graph(ctx, queries.GraphQuery{Scope: scope, Mode: queries.GraphSymbols, Name: "Save"})
	if err != nil || len(graph.Nodes) == 0 {
		t.Fatalf("graph lookup %+v %v", graph, err)
	}
	refs, err := service.Graph(ctx, queries.GraphQuery{Scope: scope, Mode: queries.GraphReferences, TargetNodeID: graph.Nodes[0].ID})
	if err != nil || len(refs.Edges) == 0 {
		t.Fatalf("crossfile refs %+v %v", refs, err)
	}
	lexical, err := service.Retrieve(ctx, queries.RetrievalQuery{Scope: scope, Mode: "lexical", Query: "charges customers"})
	if err != nil || len(lexical.Hits) == 0 {
		t.Fatalf("fallback chunk lexical %+v %v", lexical, err)
	}
	bundle, err := service.Context(ctx, queries.ContextQuery{Scope: scope, ChunkID: lexical.Hits[0].Chunk.ID})
	if err != nil || len(bundle.Items) == 0 || bundle.ViewID != receipt.ViewID {
		t.Fatalf("context %+v %v", bundle, err)
	}
	offset, complete := 0, false
	for page := 0; page < 10; page++ {
		indexed, err := service.IndexEmbeddings(ctx, queries.EmbeddingCommand{Scope: scope, ModelProfile: "test", Limit: 16, Offset: offset})
		if err != nil || indexed.Processed == 0 {
			t.Fatalf("index %+v %v", indexed, err)
		}
		if indexed.Complete {
			complete = true
			break
		}
		if indexed.NextOffset <= offset {
			t.Fatal("embedding page made no progress")
		}
		offset = indexed.NextOffset
	}
	if !complete {
		t.Fatal("fixture exceeded embedding page budget")
	}
	semantic, err := service.Retrieve(ctx, queries.RetrievalQuery{Scope: scope, Mode: "semantic", Query: "billing", ModelProfile: "test"})
	if err != nil || len(semantic.Hits) == 0 {
		t.Fatalf("semantic %+v %v", semantic, err)
	}
	hybrid, err := service.Retrieve(ctx, queries.RetrievalQuery{Scope: scope, Mode: "hybrid", Query: "billing", ModelProfile: "test"})
	if err != nil || len(hybrid.Hits) == 0 {
		t.Fatalf("hybrid %+v %v", hybrid, err)
	}
	again, err := ingester.Ingest(ctx, command)
	if err != nil || again != receipt {
		t.Fatalf("retry publication %+v %v", again, err)
	}
}
