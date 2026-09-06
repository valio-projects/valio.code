package projects

import (
	"github.com/valio-projects/valio.code/internal/domain"
	"reflect"
	"testing"
)

func definition(id domain.ProjectID, roots ...domain.ProjectSourceRoot) Definition {
	for i := range roots {
		roots[i].Role = domain.RootCode
		roots[i].Version = "1"
	}
	return Definition{Project: domain.Project{ID: id, WorkspaceID: "w", Key: string(id), Name: string(id), Kind: domain.ProjectService, Status: domain.ProjectActive}, Roots: roots}
}
func manifest(t *testing.T, repo domain.RepositoryID, files ...SourceFile) Manifest {
	t.Helper()
	m, err := BuildManifest("w", repo, files)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestMembershipSharedRootsAndMultipleRepositories(t *testing.T) {
	a := definition("a", domain.ProjectSourceRoot{RepositoryID: "r1", Path: "services/a", Include: []string{"**/*.go"}, Exclude: []string{"**/*_test.go"}}, domain.ProjectSourceRoot{RepositoryID: "r2", Path: "shared"})
	b := definition("b", domain.ProjectSourceRoot{RepositoryID: "r2", Path: "shared"})
	files := []FileRef{{"r1", "services/a/main.go"}, {"r1", "services/a/deep/x.go"}, {"r1", "services/a/main_test.go"}, {"r1", "services/ab/main.go"}, {"r2", "shared/lib.go"}}
	got, err := ResolveMembership([]Definition{b, a}, files)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]domain.ProjectID{{"a"}, {"a"}, {}, {}, {"a", "b"}}
	for i, f := range files {
		if !reflect.DeepEqual(got[f], want[i]) {
			t.Errorf("%v got %v want %v", f, got[f], want[i])
		}
	}
}
func TestProjectFingerprintIgnoresUnrelatedMonorepoChanges(t *testing.T) {
	d := definition("a", domain.ProjectSourceRoot{RepositoryID: "r", Path: "a"})
	m1 := manifest(t, "r", SourceFile{"a/main.go", []byte("one")}, SourceFile{"b/main.go", []byte("one")})
	m2 := manifest(t, "r", SourceFile{"b/main.go", []byte("two")}, SourceFile{"a/main.go", []byte("one")})
	f1, err := Fingerprint(d, []Manifest{m1})
	if err != nil {
		t.Fatal(err)
	}
	f2, _ := Fingerprint(d, []Manifest{m2})
	if m1.Fingerprint == m2.Fingerprint {
		t.Fatal("repository manifest ignored changed content")
	}
	if f1 != f2 {
		t.Fatal("unrelated path invalidated project")
	}
	m3 := manifest(t, "r", SourceFile{"a/main.go", []byte("two")})
	f3, _ := Fingerprint(d, []Manifest{m3})
	if f3 == f1 {
		t.Fatal("member content did not invalidate project")
	}
	changed := 0
	for i := range m1.Shards {
		if m1.Shards[i] != m2.Shards[i] {
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("changed shards=%d", changed)
	}
}
func TestManifestOrderAndWorkspaceScoping(t *testing.T) {
	a, b := SourceFile{"a.go", []byte("a")}, SourceFile{"b.go", []byte("b")}
	m1, m2 := manifest(t, "r", a, b), manifest(t, "r", b, a)
	if m1.Fingerprint != m2.Fingerprint {
		t.Fatal("file order changed manifest")
	}
	m3, err := BuildManifest("other", "r", []SourceFile{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if m3.Entries[0].Blob.WorkspaceID == m1.Entries[0].Blob.WorkspaceID {
		t.Fatal("unscoped blob")
	}
	if m3.Fingerprint == m1.Fingerprint {
		t.Fatal("unscoped manifest")
	}
	if _, err := BuildManifest("w", "r", []SourceFile{a, a}); err == nil {
		t.Fatal("accepted duplicate path")
	}
}
func TestDefinitionAndProfileFingerprints(t *testing.T) {
	d := definition("a", domain.ProjectSourceRoot{RepositoryID: "r", Path: ".", Include: []string{"src/**", "lib/**"}})
	f1, _ := DefinitionFingerprint(d)
	d.Roots[0].Include = []string{"lib/**", "src/**", "lib/**"}
	f2, _ := DefinitionFingerprint(d)
	if f1 != f2 {
		t.Fatal("set ordering affects definition")
	}
	d.Roots[0].Exclude = []string{"**/*_test.go"}
	f3, _ := DefinitionFingerprint(d)
	if f3 == f2 {
		t.Fatal("definition change not reflected")
	}
	a := ProfileFingerprint(map[string]string{"go": "1"}, map[string]string{"depth": "5"})
	b := ProfileFingerprint(map[string]string{"go": "2"}, map[string]string{"depth": "5"})
	if a == b {
		t.Fatal("analyzer version ignored")
	}
	c := ProfileFingerprint(map[string]string{"go": "1"}, map[string]string{"depth": "6"})
	if a == c {
		t.Fatal("configuration ignored")
	}
}
func TestInvalidDefinitionsAndGlobs(t *testing.T) {
	for _, p := range []string{"../x", "/x", "C:/x", `a\b`, "a//b"} {
		d := definition("a", domain.ProjectSourceRoot{RepositoryID: "r", Path: p})
		if d.Validate() == nil {
			t.Errorf("accepted root %q", p)
		}
	}
	for _, p := range []string{"../**", "/x", "a/**b", "[", "a//b", `a\b`} {
		if ValidateGlob(p) == nil {
			t.Errorf("accepted glob %q", p)
		}
	}
	d := definition("a", domain.ProjectSourceRoot{RepositoryID: "r", Path: "."})
	if _, err := Fingerprint(d, nil); err == nil {
		t.Fatal("accepted missing repository")
	}
}

func TestFingerprintMultipleRepositoriesAndExcludedFiles(t *testing.T) {
	d := definition("a", domain.ProjectSourceRoot{RepositoryID: "app", Path: ".", Exclude: []string{"**/*_test.go"}}, domain.ProjectSourceRoot{RepositoryID: "shared", Path: "lib"})
	app := manifest(t, "app", SourceFile{"main.go", []byte("app")}, SourceFile{"main_test.go", []byte("test")})
	shared := manifest(t, "shared", SourceFile{"lib/api.go", []byte("v1")})
	first, err := Fingerprint(d, []Manifest{app, shared})
	if err != nil {
		t.Fatal(err)
	}
	app = manifest(t, "app", SourceFile{"main.go", []byte("app")}, SourceFile{"main_test.go", []byte("changed")})
	second, err := Fingerprint(d, []Manifest{shared, app})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("excluded file or repository ordering invalidated project")
	}
	shared = manifest(t, "shared", SourceFile{"lib/api.go", []byte("v2")})
	third, err := Fingerprint(d, []Manifest{app, shared})
	if err != nil {
		t.Fatal(err)
	}
	if third == first {
		t.Fatal("second repository's source change did not invalidate project")
	}
}

func TestWorkspaceUniqueKeysAndSourceRootMetadata(t *testing.T) {
	a := definition("a", domain.ProjectSourceRoot{RepositoryID: "r", Path: "."})
	b := definition("b", domain.ProjectSourceRoot{RepositoryID: "r", Path: "."})
	b.Project.Key = a.Project.Key
	if _, err := ResolveMembership([]Definition{a, b}, nil); err == nil {
		t.Fatal("accepted duplicate workspace key")
	}
	b.Project.WorkspaceID = "other"
	if _, err := ResolveMembership([]Definition{a, b}, nil); err != nil {
		t.Fatal("key must be scoped to workspace", err)
	}
	a.Roots[0].Role = ""
	before, _ := DefinitionFingerprint(a)
	a.Roots[0].Role = domain.RootCode
	after, _ := DefinitionFingerprint(a)
	if before != after {
		t.Fatal("default role differs from code")
	}
	a.Roots[0].BuildUnit = "go-module"
	next, _ := DefinitionFingerprint(a)
	if next == after {
		t.Fatal("build unit missing from fingerprint")
	}
	a.Roots[0].Version = "2"
	last, _ := DefinitionFingerprint(a)
	if last == next {
		t.Fatal("root version missing from fingerprint")
	}
}
