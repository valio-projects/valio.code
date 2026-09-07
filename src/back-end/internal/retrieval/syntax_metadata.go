package retrieval

import (
	"strings"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

func (b Builder) syntaxChunk(source Source, start, end int, kind domainretrieval.ChunkKind, parent string, split, fallback bool, declaration syntaxfacts.Declaration) domainretrieval.Chunk {
	chunk := b.chunk(source, start, end, kind, parent, split, fallback, declaration.Kind, syntaxSignature(declaration), "")
	metadata := []string{"syntax_kind=" + declaration.Kind}
	if declaration.Visibility != "" && declaration.Visibility != "unknown" {
		metadata = append(metadata, "visibility="+declaration.Visibility)
	}
	if declaration.Type != nil {
		metadata = append(metadata, "type="+headerValue(*declaration.Type))
	}
	if declaration.UnderlyingType != nil {
		metadata = append(metadata, "underlying_type="+headerValue(*declaration.UnderlyingType))
	}
	if len(declaration.Modifiers) > 0 {
		metadata = append(metadata, "modifiers="+headerValue(strings.Join(declaration.Modifiers, ",")))
	}
	if len(declaration.Attributes) > 0 {
		metadata = append(metadata, "attributes="+headerValue(strings.Join(declaration.Attributes, ",")))
	}
	chunk.Header += "; " + strings.Join(metadata, "; ")
	chunk.Representations = syntaxRepresentations(chunk, declaration)
	return chunk
}

func syntaxRepresentations(chunk domainretrieval.Chunk, declaration syntaxfacts.Declaration) []domainretrieval.Representation {
	return representations(chunk, syntaxSignature(declaration), "")
}

func syntaxSignature(declaration syntaxfacts.Declaration) string {
	value := declaration.Name
	if len(declaration.Parameters) > 0 {
		parameters := make([]string, 0, len(declaration.Parameters))
		for _, parameter := range declaration.Parameters {
			part := parameter.Name
			if parameter.Type != nil {
				part += ": " + *parameter.Type
			}
			parameters = append(parameters, part)
		}
		value += "(" + strings.Join(parameters, ", ") + ")"
	}
	if declaration.Type != nil && (declaration.Kind == "function" || declaration.Kind == "method") {
		value += " -> " + *declaration.Type
	}
	if declaration.UnderlyingType != nil && declaration.Kind == "enum" {
		value += " : " + *declaration.UnderlyingType
	}
	return value
}

func headerValue(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ", ";", ",").Replace(value)
}
