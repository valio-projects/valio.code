package snapshots

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/codegraph"
	"github.com/valio-projects/valio.code/internal/graph"
)

func TestBuildGraphShardsScopesPackageNodesToCompatibleSource(t *testing.T) {
	view := View{Files: []FileRef{
		{ID: "repo-a-z", RepositoryID: domain.RepositoryID("repo-a"), Path: "pkg/z.go", Language: "go", ProjectIDs: []string{"project-a"}},
		{ID: "repo-a-a", RepositoryID: domain.RepositoryID("repo-a"), Path: "pkg/a.go", Language: "go", ProjectIDs: []string{"project-a"}},
		{ID: "repo-b-a", RepositoryID: domain.RepositoryID("repo-b"), Path: "pkg/a.go", Language: "go", ProjectIDs: []string{"project-b"}},
	}}
	contents := map[string]string{
		"repo-a-z": "package pkg\nfunc Z() {}\n",
		"repo-a-a": "package pkg\nfunc A() {}\n",
		"repo-b-a": "package pkg\nfunc B() {}\n",
	}

	shards, err := buildGraphShards(context.Background(), view, contents)
	if err != nil {
		t.Fatal(err)
	}
	if len(shards) != len(view.Files) {
		t.Fatalf("shard count = %d, want %d", len(shards), len(view.Files))
	}
	packageOwner := map[string]string{}
	for _, file := range view.Files {
		var shard codegraph.Graph
		if err := json.Unmarshal(shards[file.ID], &shard); err != nil {
			t.Fatal(err)
		}
		for _, node := range shard.Nodes {
			if node.Kind != codegraph.NodePackage {
				continue
			}
			if node.RepositoryID != string(file.RepositoryID) || !includesProjects(file.ProjectIDs, node.ProjectIDs) {
				t.Fatalf("package node %+v leaked into shard %+v", node, file)
			}
			packageOwner[node.RepositoryID+":"+node.ProjectIDs[0]] = file.ID
		}
	}
	if packageOwner["repo-a:project-a"] != "repo-a-a" {
		t.Fatalf("repo-a package owner = %q, want deterministic repo-a-a", packageOwner["repo-a:project-a"])
	}
	if packageOwner["repo-b:project-b"] != "repo-b-a" {
		t.Fatalf("repo-b package owner = %q", packageOwner["repo-b:project-b"])
	}
}

func TestCompatibleGraphShardRejectsIncompatibleContext(t *testing.T) {
	_, ok := compatibleGraphShard([]graph.Source{{ID: "file", RepositoryID: "repo-a", ProjectIDs: []string{"project-a"}}}, codegraph.Node{ID: "package", RepositoryID: "repo-b", ProjectIDs: []string{"project-b"}})
	if ok {
		t.Fatal("incompatible repository and project context received a shard")
	}
}
