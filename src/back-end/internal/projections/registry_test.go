package projections

import "testing"

func TestDefaultRegistryTopologicalOrder(t *testing.T) {
	r := DefaultRegistry()
	order, err := r.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 10 {
		t.Fatal("expected ten families")
	}
	positions := map[Family]int{}
	for i, f := range order {
		positions[f] = i
	}
	for _, c := range r.Contracts {
		for _, d := range c.Dependencies {
			if positions[d] >= positions[c.Family] {
				t.Fatalf("%s before dependency %s", c.Family, d)
			}
		}
	}
}
func TestRegistryRejectsCyclesAndUnknownDependencies(t *testing.T) {
	for _, r := range []Registry{{Contracts: []Contract{{SourceText, "1", []Family{Context}}, {Context, "1", []Family{SourceText}}}}, {Contracts: []Contract{{SourceText, "1", []Family{"missing"}}}}, {Contracts: []Contract{{SourceText, "1", nil}, {SourceText, "2", nil}}}} {
		if r.Validate() == nil {
			t.Fatal("accepted invalid registry")
		}
	}
}
func TestDownstreamAndProfileInvalidation(t *testing.T) {
	r := DefaultRegistry()
	got, err := r.Downstream(HistoryChange)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != HistoryChange || got[1] != Context || got[2] != Vector {
		t.Fatalf("unexpected downstream %v", got)
	}
	profiles := map[Family]string{}
	for _, c := range r.Contracts {
		profiles[c.Family] = "analyzer-v1"
	}
	before, err := r.Fingerprints("source-1", profiles)
	if err != nil {
		t.Fatal(err)
	}
	profiles[HistoryChange] = "analyzer-v2"
	after, _ := r.Fingerprints("source-1", profiles)
	for f, old := range before {
		wantChange := f == HistoryChange || f == Context || f == Vector
		if (old != after[f]) != wantChange {
			t.Errorf("unexpected profile invalidation: %s", f)
		}
	}
}
func TestReadinessKnownDenominator(t *testing.T) {
	r := DefaultRegistry()
	got, err := r.Readiness(map[Family]Status{SourceText: Ready, Structure: Partial, DataFlow: Unsupported, Vector: ExcludedByPolicy})
	if err != nil {
		t.Fatal(err)
	}
	if got.Known != 10 || got.Ready != 1 || got.Requested != 4 || got.Counts[NotRequested] != 6 {
		t.Fatalf("unexpected readiness %+v", got)
	}
	if _, err := r.Readiness(map[Family]Status{SourceText: "complete"}); err == nil {
		t.Fatal("invalid status accepted")
	}
}
