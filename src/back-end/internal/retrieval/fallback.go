package retrieval

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

func (b Builder) buildFallback(ctx context.Context, source Source, start, end int, label, parent string, split bool) ([]domainretrieval.Chunk, error) {
	max := b.maxBytes()
	if end-start <= max {
		return []domainretrieval.Chunk{b.chunk(source, start, end, domainretrieval.ChunkFileFallback, parent, split, split, label, "", "")}, nil
	}
	result := []domainretrieval.Chunk{}
	for at := start; at < end; {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		next := at + max
		if next >= end {
			next = end
		} else {
			next = safeBoundary(source.Content, next)
			if next <= at {
				next = at + max
				for next > at && !utf8.RuneStart(source.Content[next]) {
					next--
				}
			}
		}
		result = append(result, b.chunk(source, at, next, domainretrieval.ChunkFallback, parent, true, true, label, "", ""))
		at = next
	}
	return result, nil
}

func safeBoundary(text string, at int) int {
	if at >= len(text) {
		return len(text)
	}
	if at <= 0 {
		return 0
	}
	for at > 0 && !utf8.RuneStart(text[at]) {
		at--
	}
	if line := strings.LastIndexByte(text[:at], '\n'); line >= 0 && line+1 > at-512 {
		return line + 1
	}
	return at
}

func (b Builder) chunk(source Source, start, end int, kind domainretrieval.ChunkKind, parent string, split, fallback bool, label, declaration, doc string) domainretrieval.Chunk {
	text := ""
	if start != end {
		text = source.Content[start:end]
	}
	id := chunkID(source, start, end, kind, parent)
	header := fmt.Sprintf("language=%s; path=%s; kind=%s", source.Language, source.Path, label)
	if declaration != "" {
		header += "; declaration=" + declaration
	}
	if parent != "" {
		header += "; parent=" + parent
	}
	if fallback {
		header += "; split=fallback"
	} else if split {
		header += "; split=syntax"
	}
	chunk := domainretrieval.Chunk{ID: id, FileID: source.ID, RepositoryID: source.RepositoryID, ProjectIDs: append([]string(nil), source.ProjectIDs...), Start: start, End: end, Text: text, Header: header, ParentID: parent, Kind: kind, Split: split, Fallback: fallback, Representations: []domainretrieval.Representation{}}
	chunk.Representations = representations(chunk, declaration, doc)
	return chunk
}
