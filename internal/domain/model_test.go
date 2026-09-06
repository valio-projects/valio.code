package domain

import "testing"

func TestPortablePaths(t *testing.T) {
	for _, p := range []string{"", "../a", "a/../b", "/a", "C:/repo", `C:\repo`, `a\b`, "//server/share", "a//b", "a/./b", "CON", "dir/NUL.txt", "a.", "x/ ", "a\x00b", "*.go"} {
		if err := ValidateRelativePath(p, false); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
	for _, p := range []string{"main.go", "a/b.go", ".github/workflows/test.yml", "src/日本.go"} {
		if err := ValidateRelativePath(p, false); err != nil {
			t.Errorf("rejected %q: %v", p, err)
		}
	}
	if ValidateRelativePath(".", true) != nil || ValidateRelativePath(".", false) == nil {
		t.Fatal("root handling")
	}
}

func TestEvidenceIndependentDimensionsAndRange(t *testing.T) {
	e := Evidence{ID: "e", ViewID: "view", Origin: OriginHuman, Resolution: ResolutionRejected, Assertion: AssertionReported, Range: SourceRange{RepositoryID: "r", Path: "main.go", Encoding: EncodingUTF16, Start: Position{1, 2}, End: Position{1, 2}}}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
	e.Range.End.Character = 1
	if e.Validate() == nil {
		t.Fatal("accepted reversed range")
	}
	e.Range.End.Character = 3
	e.Range.Encoding = "unknown"
	if e.Validate() == nil {
		t.Fatal("accepted unknown encoding")
	}
}
func TestViewRequiresPinnedInputs(t *testing.T) {
	if (AnalysisView{ID: "v", WorkspaceID: "w"}).Validate() == nil {
		t.Fatal("accepted unpinned view")
	}
}

func validView() AnalysisView {
	return AnalysisView{ID: "view", WorkspaceID: "w", SourceSnapshotID: "composite", ProfileFingerprint: "profile-v1",
		Projects:              []ProjectRevisionRef{{WorkspaceID: "w", ProjectID: "p1", ProjectRevisionID: "p1-r1"}, {WorkspaceID: "w", ProjectID: "p2", ProjectRevisionID: "p2-r3"}},
		Repositories:          []RepositorySnapshot{{WorkspaceID: "w", SnapshotID: "capture-a", RepositoryID: "a", ManifestFingerprint: "manifest-a"}, {WorkspaceID: "w", SnapshotID: "capture-b", RepositoryID: "b", ManifestFingerprint: "manifest-b"}},
		ProjectionGenerations: []ProjectionGenerationRef{{WorkspaceID: "w", Family: "source_text", GenerationID: "gen-1"}, {WorkspaceID: "w", Family: "structure", GenerationID: "gen-2"}},
	}
}
func TestWorkspaceViewPinsMultipleProjectsAndCaptures(t *testing.T) {
	v := validView()
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*AnalysisView){
		"duplicate project":    func(v *AnalysisView) { v.Projects[1] = v.Projects[0] },
		"duplicate repository": func(v *AnalysisView) { v.Repositories[1] = v.Repositories[0] },
		"duplicate family":     func(v *AnalysisView) { v.ProjectionGenerations[1] = v.ProjectionGenerations[0] },
		"project workspace":    func(v *AnalysisView) { v.Projects[1].WorkspaceID = "other" },
		"repository workspace": func(v *AnalysisView) { v.Repositories[1].WorkspaceID = "other" },
		"projection workspace": func(v *AnalysisView) { v.ProjectionGenerations[1].WorkspaceID = "other" },
		"missing revision":     func(v *AnalysisView) { v.Projects[1].ProjectRevisionID = "" },
		"missing capture":      func(v *AnalysisView) { v.Repositories[1].SnapshotID = "" },
		"missing generation":   func(v *AnalysisView) { v.ProjectionGenerations[1].GenerationID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			v := validView()
			mutate(&v)
			if v.Validate() == nil {
				t.Fatal("accepted invalid view")
			}
		})
	}
}

func TestProjectKeyStableAcrossRename(t *testing.T) {
	p := Project{ID: "p", WorkspaceID: "w", Key: "service-a", Name: "Old name", Kind: ProjectService, Status: ProjectActive, Tags: []string{"backend"}, BuildProfiles: []string{"go-release"}, EnvironmentProfiles: []string{"production"}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	next := p
	next.Name = "New name"
	if err := next.ValidateUpdate(p); err != nil {
		t.Fatal(err)
	}
	next.Key = "new-key"
	if next.ValidateUpdate(p) == nil {
		t.Fatal("accepted key change during rename")
	}
	for _, key := range []string{"", "Service-A", "service/a", "has space", "-start", "end-", "../x"} {
		if ValidateProjectKey(key) == nil {
			t.Errorf("accepted key %q", key)
		}
	}
	p.Tags = []string{"backend", "backend"}
	if p.Validate() == nil {
		t.Fatal("accepted duplicate tags")
	}
}
func TestSourceRootRoles(t *testing.T) {
	for _, role := range []SourceRootRole{"", RootCode, RootTests, RootContracts, RootConfiguration, RootDeployment, RootDocumentation, RootGenerated} {
		r := ProjectSourceRoot{RepositoryID: "r", Path: ".", Version: "1", Role: role}
		if err := r.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	if (ProjectSourceRoot{RepositoryID: "r", Path: ".", Version: "1", Role: "vendor"}).Validate() == nil {
		t.Fatal("accepted unregistered role")
	}
}
