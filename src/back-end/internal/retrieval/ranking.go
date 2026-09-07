package retrieval

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

// Channel identifies a ranking source whose rank and score are retained.
type Channel string

const (
	// ChannelLexical identifies BM25 lexical ranking.
	ChannelLexical Channel = "lexical"
	// ChannelStructural identifies syntax-fingerprint candidate ranking.
	ChannelStructural Channel = "structural"
	// ChannelReranker identifies an optional reranking provider.
	ChannelReranker Channel = "reranker"
)

// ChannelScore records one channel's raw score and one-based rank.
type ChannelScore struct {
	// Channel identifies the ranking source.
	Channel Channel `json:"channel"`
	// Score is meaningful only within Channel and is never compared across channels.
	Score float64 `json:"score"`
	// Rank is the one-based result position within Channel.
	Rank int `json:"rank"`
}

// ScoredChunk is a chunk with fusion score and per-channel provenance.
type ScoredChunk struct {
	// Chunk is the source-grounded retrieval record.
	Chunk domainretrieval.Chunk `json:"chunk"`
	// Score is a channel score or a derived fusion score.
	Score float64 `json:"score"`
	// Provenance records all channels that contributed to Score.
	Provenance []ChannelScore `json:"provenance"`
}

// RankOptions narrows ranking to explicit source scope.
type RankOptions struct {
	// ProjectIDs selects chunks that belong to at least one listed project.
	ProjectIDs []string
	// RepositoryIDs selects chunks from listed repositories.
	RepositoryIDs []string
}

// BM25Options configures bounded lexical ranking. Zero values select the
// standard BM25 constants and an 8192-byte query limit.
type BM25Options struct {
	// Scope narrows chunks before tokenization and scoring.
	Scope RankOptions
	// MaxQueryBytes bounds the UTF-8 query input before ranking.
	MaxQueryBytes int
	// K1 controls term-frequency saturation; zero uses 1.2.
	K1 float64
	// B controls document-length normalization; zero uses 0.75.
	B float64
}

// RankBM25 validates query, scope, limits, and BM25 configuration before
// returning ranked scoped chunks. Scores are meaningful only in the lexical channel.
func RankBM25(query string, chunks []domainretrieval.Chunk, options BM25Options) ([]ScoredChunk, error) {
	maxQuery, k1, b, err := validatedBM25(options)
	if err != nil {
		return nil, err
	}
	if query == "" {
		return []ScoredChunk{}, nil
	}
	if len(query) > maxQuery {
		return nil, fmt.Errorf("retrieval query exceeds configured byte limit")
	}
	return lexicalRank(query, chunks, options.Scope, k1, b), nil
}

// LexicalRank ranks searchable chunks using BM25 over Unicode-aware identifier
// tokens. Empty queries, grouping parents, and out-of-scope chunks produce no hit.
func LexicalRank(query string, chunks []domainretrieval.Chunk, options RankOptions) []ScoredChunk {
	result, err := RankBM25(query, chunks, BM25Options{Scope: options})
	if err != nil {
		return []ScoredChunk{}
	}
	return result
}

func lexicalRank(query string, chunks []domainretrieval.Chunk, options RankOptions, k1, b float64) []ScoredChunk {
	queryTerms := tokens(query)
	if len(queryTerms) == 0 {
		return []ScoredChunk{}
	}
	docs := []domainretrieval.Chunk{}
	for _, chunk := range chunks {
		if chunk.Text != "" && inScope(chunk, options) {
			docs = append(docs, chunk)
		}
	}
	if len(docs) == 0 {
		return []ScoredChunk{}
	}
	frequencies := make([]map[string]int, len(docs))
	documentFrequency := map[string]int{}
	length := 0
	for i, chunk := range docs {
		terms := tokens(chunk.Header + "\n" + chunk.Text)
		frequencies[i] = map[string]int{}
		for _, term := range terms {
			frequencies[i][term]++
		}
		length += len(terms)
		for term := range frequencies[i] {
			documentFrequency[term]++
		}
	}
	averageLength := float64(length) / float64(len(docs))
	if averageLength == 0 {
		return []ScoredChunk{}
	}
	result := []ScoredChunk{}
	for i, chunk := range docs {
		score := 0.0
		for _, term := range queryTerms {
			frequency := frequencies[i][term]
			if frequency == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(docs)-documentFrequency[term])+0.5)/(float64(documentFrequency[term])+0.5))
			denominator := float64(frequency) + k1*(1-b+b*float64(sum(frequencies[i]))/averageLength)
			score += idf * (float64(frequency) * (k1 + 1)) / denominator
		}
		if score > 0 {
			result = append(result, ScoredChunk{Chunk: chunk, Score: score})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Chunk.ID < result[j].Chunk.ID
	})
	for i := range result {
		result[i].Provenance = []ChannelScore{{Channel: ChannelLexical, Score: result[i].Score, Rank: i + 1}}
	}
	return result
}

func validatedBM25(options BM25Options) (int, float64, float64, error) {
	maxQuery := options.MaxQueryBytes
	if maxQuery == 0 {
		maxQuery = 8192
	}
	if maxQuery < 1 || maxQuery > 64*1024 {
		return 0, 0, 0, fmt.Errorf("retrieval query byte limit must be between 1 and 65536")
	}
	for _, values := range [][]string{options.Scope.ProjectIDs, options.Scope.RepositoryIDs} {
		if len(values) > 128 {
			return 0, 0, 0, fmt.Errorf("retrieval scope contains too many identifiers")
		}
		seen := map[string]bool{}
		for _, value := range values {
			if value == "" || strings.TrimSpace(value) != value || seen[value] {
				return 0, 0, 0, fmt.Errorf("retrieval scope identifiers must be unique and trimmed")
			}
			seen[value] = true
		}
	}
	k1, b := options.K1, options.B
	if k1 == 0 {
		k1 = 1.2
	}
	if b == 0 {
		b = 0.75
	}
	if math.IsNaN(k1) || math.IsInf(k1, 0) || k1 <= 0 || math.IsNaN(b) || math.IsInf(b, 0) || b < 0 || b > 1 {
		return 0, 0, 0, fmt.Errorf("invalid BM25 configuration")
	}
	return maxQuery, k1, b, nil
}

// FuseRRF combines channels by reciprocal rank fusion. k defaults to 60 and
// preserves each contributing rank and raw score in the returned provenance.
func FuseRRF(k int, channels ...[]ScoredChunk) []ScoredChunk {
	if k <= 0 {
		k = 60
	}
	byID := map[string]ScoredChunk{}
	seenContribution := map[string]bool{}
	for _, channel := range channels {
		seen := map[string]bool{}
		for index, item := range channel {
			if item.Chunk.ID == "" || seen[item.Chunk.ID] {
				continue
			}
			seen[item.Chunk.ID] = true
			entry := byID[item.Chunk.ID]
			if entry.Chunk.ID == "" {
				entry.Chunk = item.Chunk
				entry.Provenance = []ChannelScore{}
			}
			provenance := item.Provenance
			if len(provenance) == 0 {
				provenance = []ChannelScore{{Channel: ChannelLexical, Score: item.Score, Rank: index + 1}}
			}
			for _, proof := range provenance {
				rank := proof.Rank
				if rank <= 0 {
					rank = index + 1
					proof.Rank = rank
				}
				key := item.Chunk.ID + "\x00" + string(proof.Channel)
				if seenContribution[key] {
					continue
				}
				seenContribution[key] = true
				entry.Score += 1 / float64(k+rank)
				entry.Provenance = append(entry.Provenance, proof)
			}
			byID[item.Chunk.ID] = entry
		}
	}
	result := make([]ScoredChunk, 0, len(byID))
	for _, item := range byID {
		sort.Slice(item.Provenance, func(i, j int) bool { return item.Provenance[i].Channel < item.Provenance[j].Channel })
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Chunk.ID < result[j].Chunk.ID
	})
	return result
}

func inScope(chunk domainretrieval.Chunk, options RankOptions) bool {
	if len(options.RepositoryIDs) > 0 && !contains(options.RepositoryIDs, chunk.RepositoryID) {
		return false
	}
	if len(options.ProjectIDs) == 0 {
		return true
	}
	for _, project := range chunk.ProjectIDs {
		if contains(options.ProjectIDs, project) {
			return true
		}
	}
	return false
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func sum(values map[string]int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func tokens(text string) []string {
	terms := []string{}
	var word []rune
	flush := func() {
		if len(word) == 0 {
			return
		}
		value := strings.ToLower(string(word))
		terms = append(terms, value)
		for i := 1; i < len(word); i++ {
			if unicode.IsLower(word[i-1]) && unicode.IsUpper(word[i]) {
				terms = append(terms, strings.ToLower(string(word[:i])), strings.ToLower(string(word[i:])))
			}
		}
		word = nil
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word = append(word, r)
		} else {
			flush()
		}
	}
	flush()
	return terms
}
