package search

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func run(t *testing.T, files []File, q string, o Options) Result {
	t.Helper()
	r, e := Search(context.Background(), MemorySource(files), q, o)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func ids(r Result) []string {
	out := []string{}
	for _, m := range r.Matches {
		out = append(out, m.FileID)
	}
	return out
}
func TestBooleanFileSemantics(t *testing.T) {
	f := []File{{ID: "a", Content: "alpha\nbeta"}, {ID: "b", Content: "alpha gamma"}, {ID: "c", Content: "gamma"}, {ID: "d", Content: "beta"}}
	for _, tc := range []struct {
		q    string
		want []string
	}{
		{"alpha beta", []string{"a"}}, {"alpha NOT beta", []string{"b"}}, {"alpha OR gamma AND NOT beta", []string{"a", "b", "c"}}, {"(alpha OR gamma) AND NOT beta", []string{"b", "c"}}, {"NOT (alpha OR gamma)", []string{"d"}}, {"NOT NOT alpha", []string{"a", "b"}},
	} {
		t.Run(tc.q, func(t *testing.T) {
			got := ids(run(t, f, tc.q, Options{}))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
func TestErrors(t *testing.T) {
	for _, q := range []string{"", "alpha AND", "OR alpha", "()", "(alpha", "alpha)", "owner:me", "test:maybe", "generated:/true/", `"abc`, `content:/(/`, `"bad\q"`, `"x"y`} {
		t.Run(q, func(t *testing.T) {
			_, err := Search(context.Background(), MemorySource{}, q, Options{})
			if err == nil {
				t.Fatalf("expected error for %q", q)
			}
		})
	}
}
func TestQuotedAndPunctuation(t *testing.T) {
	text := "say \"hi\"\\next\r\nfoo.bar!"
	r := run(t, []File{{ID: "1", Content: text}}, `"say \"hi\"\\next\r\nfoo.bar!"`, Options{})
	if r.Total != 1 || r.Matches[0].Ranges[0] != (Range{0, len(text)}) {
		t.Fatalf("bad result %+v", r)
	}
}
func TestRegexConservativeCandidates(t *testing.T) {
	f := []File{{ID: "a", Content: "needle"}, {ID: "b", Content: "other"}, {ID: "c", Content: "prefixYYend"}, {ID: "d", Content: "end"}}
	for _, tc := range []struct {
		q    string
		want []string
	}{
		{`/needle|other/`, []string{"a", "b"}}, {`/(prefix.*)?end/`, []string{"c", "d"}}, {`/(prefix.*|)end/`, []string{"c", "d"}}, {`NOT /needle|other/`, []string{"c", "d"}}, {`/(?:prefix){0,2}.*end/`, []string{"c", "d"}},
	} {
		if got := ids(run(t, f, tc.q, Options{})); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s got %v want %v", tc.q, got, tc.want)
		}
	}
}
func TestCrossBlockAndLongLiteral(t *testing.T) {
	for _, literal := range []string{"boundary-needle", strings.Repeat("long-🙂", 5000)} {
		text := strings.Repeat("x", BlockBytes-5) + literal + "tail"
		r := run(t, []File{{ID: "f", Content: text}}, `"`+literal+`"`, Options{})
		if r.Total != 1 {
			t.Fatalf("lost literal with %d bytes", len(literal))
		}
		want := Range{BlockBytes - 5, BlockBytes - 5 + len(literal)}
		if got := r.Matches[0].Ranges[0]; got != want {
			t.Fatalf("got %+v want %+v", got, want)
		}
	}
	idx := BuildIndex(strings.Repeat("x", BlockBytes-1) + "αβγdef")
	for _, tri := range trigrams(fold("xαβγdef")) {
		if _, ok := idx.Trigrams[tri]; !ok {
			t.Fatalf("missing cross-boundary trigram %s", tri)
		}
	}
}
func TestUnicodeCaseAndOriginalOffsets(t *testing.T) {
	text := "🙂 KELVIN Σςσ\r\n"
	for _, q := range []string{`"kelvin σσσ"`, `/kelvin σσσ/`} {
		r := run(t, []File{{Content: text}}, q, Options{})
		if r.Total != 1 {
			t.Fatalf("lost Unicode fold: %s", q)
		}
		got := r.Matches[0].Ranges[0]
		if text[got.Start:got.End] != "KELVIN Σςσ" {
			t.Fatalf("invalid source range %+v", got)
		}
	}
	if run(t, []File{{Content: text}}, `kelvin`, Options{CaseSensitive: true}).Total != 0 {
		t.Fatal("case sensitive ignored")
	}
}
func TestProjectScopeSnapshotAndPaging(t *testing.T) {
	f := []File{{ID: "3", Content: "yes", ProjectIDs: []string{"p", "private"}, RepositoryID: "r", SnapshotID: "s2"}, {ID: "2", Content: "yes", ProjectIDs: []string{"private"}, RepositoryID: "r", SnapshotID: "s1"}, {ID: "1", Content: "yes", ProjectIDs: []string{"p", "private"}, RepositoryID: "r", SnapshotID: "s1"}, {ID: "0", Content: "yes", ProjectIDs: []string{"p"}, RepositoryID: "r", SnapshotID: "s1"}}
	r := run(t, f, "yes", Options{ProjectIDs: []string{"p"}, SnapshotIDs: []string{"s1"}, Limit: 1, Offset: 1})
	if r.Total != 2 || !r.Complete || r.Truncated || !reflect.DeepEqual(ids(r), []string{"1"}) {
		t.Fatalf("%+v", r)
	}
	if !reflect.DeepEqual(r.Matches[0].ProjectIDs, []string{"p"}) {
		t.Fatal("leaked project context")
	}
	if run(t, f, "project:private", Options{ProjectIDs: []string{"p"}}).Total != 0 {
		t.Fatal("project filter escaped selected context")
	}
}
func TestBoundedShortScanAndNoEarlyTopK(t *testing.T) {
	f := []File{{ID: "1", Content: "abc bca cab"}, {ID: "2", Content: "abcabc"}, {ID: "3", Content: "abcabc"}}
	r := run(t, f, "abcabc", Options{Limit: 1})
	if r.Total != 2 || r.Matches[0].FileID != "2" || !r.Complete {
		t.Fatalf("verification must precede paging: %+v", r)
	}
	r = run(t, f, "a", Options{MaxScanFiles: 1})
	if r.Total != 1 || r.Complete || !r.Truncated || r.ScannedFiles != 1 {
		t.Fatalf("short scan missing truncation: %+v", r)
	}
	r = run(t, f, "a", Options{MaxScanBytes: 1})
	if r.Total != 0 || r.Complete || !r.Truncated {
		t.Fatalf("byte limit ignored: %+v", r)
	}
}

func TestRangeBudgetKeepsMatchedFileTotalsExact(t *testing.T) {
	content := strings.Repeat("a", maxRangesPerMatch+1)
	for _, query := range []string{"a", `/a/`, "a AND a"} {
		r := run(t, []File{{ID: "dense", Content: content}, {ID: "other", Content: "a"}}, query, Options{})
		if r.Total != 2 || !r.Complete || !r.Truncated || len(r.Matches) != 2 {
			t.Fatalf("%q lost exact file result: %+v", query, r)
		}
		if len(r.Matches[0].Ranges) != maxRangesPerMatch || !r.Matches[0].RangesTruncated {
			t.Fatalf("%q did not retain bounded range sample: %+v", query, r.Matches[0])
		}
		if r.Matches[1].RangesTruncated {
			t.Fatalf("%q marked a small result truncated: %+v", query, r.Matches[1])
		}
	}
}

func TestRangeBudgetDoesNotChangeNotTruth(t *testing.T) {
	dense := strings.Repeat("a", maxRangesPerMatch+1)
	r := run(t, []File{{ID: "dense", Content: dense}, {ID: "clean", Content: "b"}}, "NOT a", Options{})
	if r.Total != 1 || !r.Complete || r.Truncated || !reflect.DeepEqual(ids(r), []string{"clean"}) {
		t.Fatalf("NOT must remain exact without irrelevant ranges: %+v", r)
	}
}

func TestRangeBudgetCapsSymbolOccurrences(t *testing.T) {
	symbols := make([]Symbol, maxRangesPerMatch+1)
	for i := range symbols {
		symbols[i] = Symbol{Name: "needle", Kind: "function", Start: i, End: i + 1}
	}
	r := run(t, []File{{ID: "symbols", Content: strings.Repeat("x", maxRangesPerMatch+1), Symbols: symbols}}, "symbol:needle", Options{})
	if r.Total != 1 || !r.Complete || !r.Truncated || !r.Matches[0].RangesTruncated || len(r.Matches[0].Ranges) != maxRangesPerMatch {
		t.Fatalf("symbol range cap lost result truth: %+v", r)
	}
}
func TestFilters(t *testing.T) {
	f := []File{{ID: "f", Path: "src/a_test.go", Content: "func Hello() {}", Language: "go", RepositoryID: "repo", ProjectIDs: []string{"p"}, Test: true, Symbols: []Symbol{{Name: "Hello", Kind: "function", Start: 5, End: 10}}}}
	r := run(t, f, `project:p repo:repo file:a_test.go lang:go path:src symbol:Hello kind:function test:true generated:false`, Options{})
	if r.Total != 1 {
		t.Fatalf("%+v", r)
	}
	if got := r.Matches[0].Ranges; !reflect.DeepEqual(got, []Range{{5, 10}}) {
		t.Fatalf("%+v", got)
	}
}
func TestExactAndRegexModes(t *testing.T) {
	f := []File{{ID: "1", Content: "Alpha"}, {ID: "2", Content: "Alpha Beta"}}
	if got := ids(run(t, f, "alpha", Options{Mode: Exact})); !reflect.DeepEqual(got, []string{"1"}) {
		t.Fatalf("%v", got)
	}
	if run(t, f, `"^Alpha$"`, Options{Mode: Regex}).Total != 1 {
		t.Fatal("regex mode ignored")
	}
}

func TestConfigurableComponentsPreserveCandidateSafety(t *testing.T) {
	parser := QueryParser{MaxQueryBytes: 3}
	if _, err := parser.Parse("long"); err == nil {
		t.Fatal("query byte budget ignored")
	}
	text := "ab🙂cdefΣKtail"
	index := (IndexBuilder{BlockSize: 12}).Build(text)
	for _, tri := range trigrams(fold(text)) {
		if _, ok := index.Trigrams[tri]; !ok {
			t.Fatalf("custom block builder lost %s", tri)
		}
	}
	r, err := (Engine{Source: MemorySource{{ID: "f", Content: text}}, IndexBuilder: IndexBuilder{BlockSize: 12}}).Search(context.Background(), `"cdefσk"`)
	if err != nil || r.Total != 1 {
		t.Fatalf("engine integration %+v %v", r, err)
	}
}

func TestQueryHandlerKeepsRequestScopesSeparate(t *testing.T) {
	handler := QueryHandler{Engine: Engine{Source: MemorySource{{ID: "a", Content: "yes", ProjectIDs: []string{"a"}}, {ID: "b", Content: "yes", ProjectIDs: []string{"b"}}}}}
	for _, p := range []string{"a", "b"} {
		r, err := handler.Handle(context.Background(), SearchQuery{Expression: "yes", Options: Options{ProjectIDs: []string{p}}})
		if err != nil || r.Total != 1 || r.Matches[0].FileID != p {
			t.Fatalf("request scope leaked %+v %v", r, err)
		}
	}
	if len(handler.Engine.Options.ProjectIDs) != 0 {
		t.Fatal("handler mutated shared engine")
	}
}

func FuzzCandidateNeverDropsLiteral(f *testing.F) {
	for _, s := range []string{"abc", "KΣςλ", "a\r\nb", "🙂def", strings.Repeat("x", BlockBytes-1) + "abc"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 100000 || s == "" {
			t.Skip()
		}
		idx := BuildIndex(s)
		for _, tri := range trigrams(fold(s)) {
			if _, ok := idx.Trigrams[tri]; !ok {
				t.Fatalf("index omitted %s", tri)
			}
		}
	})
}
