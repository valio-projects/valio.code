package retrieval

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"unicode"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

func representations(chunk domainretrieval.Chunk, declaration, doc string) []domainretrieval.Representation {
	result := []domainretrieval.Representation{}
	if chunk.Text != "" {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationCode, Text: chunk.Text})
	}
	if declaration != "" {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationSymbol, Text: "symbol: " + declaration + "\n" + chunk.Header})
	}
	if doc != "" {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationDocumentation, Text: doc})
	}
	if chunk.Header != "" {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationContext, Text: chunk.Header})
	}
	if errors := errorFacts(chunk.Text); len(errors) > 0 {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationError, Text: "errors: " + strings.Join(errors, ", ")})
	}
	if declaration != "" && exported(declaration) {
		result = append(result, domainretrieval.Representation{Kind: domainretrieval.RepresentationAPI, Text: "api: " + declaration + "\n" + chunk.Header})
	}
	return result
}

func exported(name string) bool {
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}

func errorFacts(text string) []string {
	if text == "" {
		return nil
	}
	source := "package retrieval\n" + text
	if !strings.HasPrefix(strings.TrimSpace(text), "func ") && !strings.HasPrefix(strings.TrimSpace(text), "type ") {
		source = "package retrieval\nfunc candidate() {\n" + text + "\n}"
	}
	file, err := parser.ParseFile(token.NewFileSet(), "chunk.go", source, parser.AllErrors)
	if err != nil || file == nil {
		return nil
	}
	facts := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Ident:
			if node.Name == "error" {
				facts["error"] = true
			}
		case *ast.CallExpr:
			switch called := node.Fun.(type) {
			case *ast.Ident:
				if called.Name == "panic" {
					facts["panic"] = true
				}
			case *ast.SelectorExpr:
				if called.Sel.Name == "Error" || called.Sel.Name == "Errorf" || called.Sel.Name == "New" {
					facts[called.Sel.Name] = true
				}
			}
		}
		return true
	})
	result := make([]string, 0, len(facts))
	for value := range facts {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
