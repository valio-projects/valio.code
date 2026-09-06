package search

import (
	"encoding/hex"
	"strings"
	"unicode"
)

const BlockBytes = 16 * 1024

type Block struct {
	Start  int
	End    int
	Tokens []string
}
type Index struct {
	Blocks   []Block
	Trigrams map[string]struct{}
}
type IndexBuilder struct{ BlockSize int }

// foldRune chooses one representative from Unicode's SimpleFold orbit, the
// same case equivalence used by Go's regexp engine (including sigma and Kelvin).
func foldRune(r rune) rune {
	lowest := r
	for x := unicode.SimpleFold(r); x != r; x = unicode.SimpleFold(x) {
		if x < lowest {
			lowest = x
		}
	}
	return lowest
}
func fold(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(foldRune(r))
	}
	return b.String()
}
func trigrams(s string) []string {
	r := []rune(s)
	var out []string
	for i := 0; i+2 < len(r); i++ {
		out = append(out, hex.EncodeToString([]byte(string(r[i:i+3]))))
	}
	return out
}

// BuildIndex makes source-byte-aligned 16 KiB blocks with two-rune overlap.
// Tokens encode normalized rune trigrams as hex so punctuation is safe for
// database tokenization. Aggregate membership is necessary, never sufficient.
func BuildIndex(content string) Index {
	return (IndexBuilder{}).Build(content)
}
func (builder IndexBuilder) Build(content string) Index {
	blockSize := builder.BlockSize
	if blockSize <= 0 {
		blockSize = BlockBytes
	}
	if blockSize < 12 {
		blockSize = 12
	}
	idx := Index{Trigrams: map[string]struct{}{}}
	type point struct {
		start, end int
		r          rune
	}
	var points []point
	for start, r := range content {
		points = append(points, point{start: start, r: r})
	}
	for i := range points {
		points[i].end = len(content)
		if i+1 < len(points) {
			points[i].end = points[i+1].start
		}
	}
	for first := 0; first < len(points); {
		last := first
		for last < len(points) && (last == first || points[last].end-points[first].start <= blockSize) {
			last++
		}
		var b strings.Builder
		for _, p := range points[first:last] {
			b.WriteRune(foldRune(p.r))
		}
		tokens := trigrams(b.String())
		for _, t := range tokens {
			idx.Trigrams[t] = struct{}{}
		}
		idx.Blocks = append(idx.Blocks, Block{Start: points[first].start, End: points[last-1].end, Tokens: tokens})
		if last == len(points) {
			break
		}
		next := last - 2
		if next <= first {
			next = last
		}
		first = next
	}
	return idx
}
