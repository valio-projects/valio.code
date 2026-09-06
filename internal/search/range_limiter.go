package search

const maxRangesPerMatch = 1000

// rangeMatches carries exact match truth and a bounded sample of source spans.
// truncated is the sentinel observed after the retained range budget.
type rangeMatches struct {
	ranges    []Range
	truncated bool
}

func (matches *rangeMatches) add(r Range) {
	if len(matches.ranges) == maxRangesPerMatch {
		matches.truncated = true
		return
	}
	matches.ranges = append(matches.ranges, r)
}

func combineRanges(left, right rangeMatches) rangeMatches {
	result := rangeMatches{truncated: left.truncated || right.truncated}
	for _, r := range left.ranges {
		result.add(r)
	}
	for _, r := range right.ranges {
		result.add(r)
	}
	return result
}
