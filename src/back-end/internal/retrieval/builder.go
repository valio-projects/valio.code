package retrieval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/domain"
	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

const (
	defaultChunkBytes = 4096
	minimumChunkBytes = 64
	maximumChunkBytes = 64 * 1024
)

// Builder creates deterministic bounded chunks. MaxBytes defaults to 4096 and
// is measured in original UTF-8 bytes. Explicit budgets must be between 64 and
// 65536 bytes so every resulting record has a practical bounded size.
type Builder struct {
	// MaxBytes limits canonical text carried by searchable chunks.
	MaxBytes int
}

// Build partitions sources in deterministic input-independent order. It never
// mutates inputs and checks context between files.
func (b Builder) Build(ctx context.Context, sources []Source) ([]domainretrieval.Chunk, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	sorted := append([]Source(nil), sources...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].RepositoryID != sorted[j].RepositoryID {
			return sorted[i].RepositoryID < sorted[j].RepositoryID
		}
		if sorted[i].Path != sorted[j].Path {
			return sorted[i].Path < sorted[j].Path
		}
		return sorted[i].ID < sorted[j].ID
	})
	seen := map[string]bool{}
	chunks := []domainretrieval.Chunk{}
	for _, source := range sorted {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if source.ID == "" || source.RepositoryID == "" || source.Path == "" {
			return nil, fmt.Errorf("retrieval source requires id, repository, and path")
		}
		if err := domain.ValidateRelativePath(source.Path, false); err != nil {
			return nil, fmt.Errorf("retrieval source path is invalid")
		}
		if !utf8.ValidString(source.Content) {
			return nil, fmt.Errorf("retrieval source must contain valid UTF-8")
		}
		if seen[source.ID] {
			return nil, fmt.Errorf("duplicate retrieval source %q", source.ID)
		}
		seen[source.ID] = true
		projects, err := normalizedProjects(source.ProjectIDs)
		if err != nil {
			return nil, err
		}
		source.ProjectIDs = projects
		if strings.EqualFold(source.Language, "go") || strings.HasSuffix(strings.ToLower(source.Path), ".go") {
			built, err := b.buildGo(ctx, source)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, built...)
		} else {
			built, err := b.buildFallback(ctx, source, 0, len(source.Content), "file", "", false)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, built...)
		}
	}
	return chunks, nil
}

func (b Builder) maxBytes() int {
	if b.MaxBytes <= 0 {
		return defaultChunkBytes
	}
	return b.MaxBytes
}

func (b Builder) validate() error {
	if b.MaxBytes < 0 || (b.MaxBytes > 0 && (b.MaxBytes < minimumChunkBytes || b.MaxBytes > maximumChunkBytes)) {
		return fmt.Errorf("retrieval chunk budget must be between %d and %d bytes", minimumChunkBytes, maximumChunkBytes)
	}
	return nil
}

func normalizedProjects(values []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || strings.TrimSpace(value) != value {
			return nil, fmt.Errorf("project identifiers must be nonempty and trimmed")
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result, nil
}

func chunkID(source Source, start, end int, kind domainretrieval.ChunkKind, parent string) string {
	payload := source.RepositoryID + "\x00" + source.ID + "\x00" + fmt.Sprintf("%d\x00%d", start, end) + "\x00" + string(kind) + "\x00" + parent
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}
