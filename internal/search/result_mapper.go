package search

import "sort"

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
func scopedProjects(f File, o Options) []string {
	var p []string
	for _, id := range f.ProjectIDs {
		if len(o.ProjectIDs) == 0 || contains(o.ProjectIDs, id) {
			p = append(p, id)
		}
	}
	sort.Strings(p)
	return compactStrings(p)
}
func compactStrings(s []string) []string {
	out := s[:0]
	for _, v := range s {
		if len(out) == 0 || v != out[len(out)-1] {
			out = append(out, v)
		}
	}
	return out
}
func canonicalRanges(rs []Range) []Range {
	sort.Slice(rs, func(i, j int) bool {
		if rs[i].Start != rs[j].Start {
			return rs[i].Start < rs[j].Start
		}
		return rs[i].End < rs[j].End
	})
	out := make([]Range, 0, len(rs))
	for _, r := range rs {
		if len(out) == 0 || out[len(out)-1] != r {
			out = append(out, r)
		}
	}
	return out
}
