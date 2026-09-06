package search

import (
	"context"
	"fmt"
	"sort"
)

// Search parses input and evaluates it against source using request-local options.
func Search(ctx context.Context, source Source, input string, o Options) (Result, error) {
	return (Engine{Source: source, Options: o}).Search(ctx, input)
}

// Engine contains reusable search dependencies and defaults.
type Engine struct {
	// Source supplies stable file snapshots.
	Source Source
	// Options apply when callers use Engine methods.
	Options Options
	// Parser parses expressions; its zero value has the default query limit.
	Parser QueryParser
	// IndexBuilder creates conservative candidate indexes.
	IndexBuilder IndexBuilder
}

// Search parses input then evaluates it through this engine.
func (engine Engine) Search(ctx context.Context, input string) (Result, error) {
	q, e := engine.Parser.Parse(input)
	if e != nil {
		return Result{}, e
	}
	return engine.Execute(ctx, q)
}

// Execute evaluates all eligible files before paging. There is no hidden top K.
// Exact mode compares complete field values; /regex/ values override Mode.
func Execute(ctx context.Context, source Source, q *Query, o Options) (Result, error) {
	return (Engine{Source: source, Options: o}).Execute(ctx, q)
}

// Execute evaluates an already parsed query through this engine.
func (engine Engine) Execute(ctx context.Context, q *Query) (Result, error) {
	source, o := engine.Source, engine.Options
	if source == nil {
		return Result{}, fmt.Errorf("search source is required")
	}
	if q == nil || q.root == nil {
		return Result{}, fmt.Errorf("query is empty")
	}
	if o.Mode != "" && o.Mode != Substring && o.Mode != Exact && o.Mode != Regex {
		return Result{}, fmt.Errorf("unsupported match mode %q", o.Mode)
	}
	if o.Offset < 0 || o.Limit < 0 || o.MaxScanFiles < 0 || o.MaxScanBytes < 0 {
		return Result{}, fmt.Errorf("limits must be nonnegative")
	}
	c, e := compile(q.root, o)
	if e != nil {
		return Result{}, e
	}
	files, e := source.Files(ctx)
	if e != nil {
		return Result{}, e
	}
	files = append([]File(nil), files...)
	sort.SliceStable(files, func(i, j int) bool {
		a, b := files[i], files[j]
		if a.RepositoryID != b.RepositoryID {
			return a.RepositoryID < b.RepositoryID
		}
		if a.SnapshotID != b.SnapshotID {
			return a.SnapshotID < b.SnapshotID
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.ID < b.ID
	})
	if o.MaxScanFiles == 0 {
		o.MaxScanFiles = 100000
	}
	if o.MaxScanBytes == 0 {
		o.MaxScanBytes = 256 << 20
	}
	if o.Limit == 0 {
		o.Limit = 100
	}
	r := Result{Complete: true, Matches: []Match{}}
	var matched []Match
	for _, f := range files {
		if e := ctx.Err(); e != nil {
			return Result{}, e
		}
		projects := scopedProjects(f, o)
		if len(o.ProjectIDs) > 0 && len(projects) == 0 {
			continue
		}
		if len(o.RepositoryIDs) > 0 && !contains(o.RepositoryIDs, f.RepositoryID) {
			continue
		}
		if len(o.SnapshotIDs) > 0 && !contains(o.SnapshotIDs, f.SnapshotID) {
			continue
		}
		if r.ScannedFiles >= o.MaxScanFiles || int64(len(f.Content)) > o.MaxScanBytes-r.ScannedBytes {
			r.Complete = false
			r.Truncated = true
			break
		}
		r.ScannedFiles++
		r.ScannedBytes += int64(len(f.Content))
		if !c.candidate(engine.IndexBuilder.Build(f.Content)) {
			continue
		}
		ok, ranges := c.evaluate(f, projects)
		if !ok {
			continue
		}
		if ranges.truncated {
			r.Truncated = true
		}
		matched = append(matched, Match{FileID: f.ID, Path: f.Path, RepositoryID: f.RepositoryID, SnapshotID: f.SnapshotID, ProjectIDs: projects, Ranges: canonicalRanges(ranges.ranges), RangesTruncated: ranges.truncated})
	}
	r.Total = len(matched)
	start := min(o.Offset, len(matched))
	end := min(start+min(o.Limit, len(matched)-start), len(matched))
	r.Matches = append(r.Matches, matched[start:end]...)
	return r, nil
}
