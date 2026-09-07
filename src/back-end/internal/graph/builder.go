package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
)

// MaxFiles is the largest input collection the builder accepts independently
// of application-level upload limits.
const MaxFiles = 10000

// MaxSourceBytes bounds total supplied source bytes before parsing.
const MaxSourceBytes = 16 << 20

// MaxNodes bounds a graph result before the builder returns a partial graph.
const MaxNodes = 200000

// MaxEdges bounds a graph result before the builder returns a partial graph.
const MaxEdges = 400000

// ErrBudget reports that an otherwise valid input exceeded a local graph limit.
var ErrBudget = errors.New("code graph traversal budget exceeded")

type parsedSource struct {
	source Source
	file   *ast.File
}

type packageGroup struct {
	repositoryID string
	directory    string
	name         string
	projectID    string
	files        []parsedSource
}

type graphBuilder struct {
	ctx         context.Context
	fset        *token.FileSet
	nodes       map[string]codegraph.Node
	edges       map[string]codegraph.Edge
	diagnostics map[string]codegraph.Diagnostic
	complete    codegraph.Completeness
	truncated   bool
}

// Build parses and type-checks only compatible Go package variants made from
// input. It never executes a build, downloads dependencies, or resolves an
// external import. Exact references, local access targets, type/member relations,
// and call targets are emitted only when go/types resolves them to a supplied local declaration.
// Reads and writes are access facts, not CFG, DFG, alias, or control-dependence analysis.
//
// Source IDs must be stable within the caller's source snapshot. Returned node
// and edge IDs are deterministic from those IDs and source ranges, not a view
// version. Build returns a partial graph with ErrBudget on local limit exhaustion
// and returns the context error when cancellation is observed.
func Build(ctx context.Context, input []Source) (codegraph.Graph, error) {
	b := graphBuilder{
		ctx:         ctx,
		fset:        token.NewFileSet(),
		nodes:       map[string]codegraph.Node{},
		edges:       map[string]codegraph.Edge{},
		diagnostics: map[string]codegraph.Diagnostic{},
		complete:    codegraph.Completeness{InputFiles: len(input)},
	}
	if err := b.validateInput(input); err != nil {
		return b.graph(), err
	}
	parsed := b.parse(input)
	if err := ctx.Err(); err != nil {
		return b.graph(), err
	}
	for _, p := range parsed {
		b.addFile(p.source)
	}
	groups := groupPackages(parsed)
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return b.graph(), err
		}
		if b.truncated {
			break
		}
		b.checkGroup(groups[key])
	}
	graph := b.graph()
	if err := graph.Validate(); err != nil {
		return graph, err
	}
	if b.truncated {
		return graph, ErrBudget
	}
	return graph, nil
}

func (b *graphBuilder) validateInput(input []Source) error {
	if len(input) > MaxFiles {
		return fmt.Errorf("%w: %d files exceeds %d", ErrBudget, len(input), MaxFiles)
	}
	seen := map[string]bool{}
	total := 0
	for _, source := range input {
		if err := b.ctx.Err(); err != nil {
			return err
		}
		if source.ID == "" || source.RepositoryID == "" || source.Path == "" {
			return errors.New("graph source requires id, repository id, and path")
		}
		if seen[source.ID] {
			return fmt.Errorf("duplicate graph source id %q", source.ID)
		}
		if strings.Contains(source.Path, "\\") || strings.HasPrefix(source.Path, "/") || path.Clean(source.Path) != source.Path || source.Path == "." || strings.HasPrefix(source.Path, "../") {
			return fmt.Errorf("graph source path %q is not repository-relative", source.Path)
		}
		if !utf8.ValidString(source.Content) {
			return fmt.Errorf("graph source %q is not valid UTF-8", source.ID)
		}
		projects := map[string]bool{}
		for _, projectID := range source.ProjectIDs {
			if projectID == "" || projects[projectID] {
				return fmt.Errorf("graph source %q has invalid project contexts", source.ID)
			}
			projects[projectID] = true
		}
		total += len(source.Content)
		if total > MaxSourceBytes {
			return fmt.Errorf("%w: source bytes exceed %d", ErrBudget, MaxSourceBytes)
		}
		seen[source.ID] = true
	}
	return nil
}

func (b *graphBuilder) parse(input []Source) []parsedSource {
	parsed := make([]parsedSource, 0, len(input))
	for _, source := range input {
		if b.ctx.Err() != nil || b.truncated {
			break
		}
		filename := source.Path + "@" + source.ID
		file, err := parser.ParseFile(b.fset, filename, source.Content, parser.AllErrors|parser.ParseComments)
		if err != nil {
			rangeValue := codegraph.ByteRange{FileID: source.ID, Start: 0, End: 0}
			b.diagnostic("parse_error", "Go syntax could not be fully parsed", &rangeValue)
		}
		if file == nil {
			continue
		}
		b.complete.ParsedFiles++
		parsed = append(parsed, parsedSource{source: copiedSource(source), file: file})
	}
	return parsed
}

func copiedSource(source Source) Source {
	copy := source
	copy.ProjectIDs = sortedStrings(source.ProjectIDs)
	return copy
}

func groupPackages(parsed []parsedSource) map[string]*packageGroup {
	groups := map[string]*packageGroup{}
	for _, item := range parsed {
		directory := path.Dir(item.source.Path)
		if directory == "." {
			directory = ""
		}
		contexts := item.source.ProjectIDs
		if len(contexts) == 0 {
			contexts = []string{""}
		}
		for _, projectID := range contexts {
			key := strings.Join([]string{item.source.RepositoryID, directory, item.file.Name.Name, projectID}, "\x00")
			group := groups[key]
			if group == nil {
				group = &packageGroup{repositoryID: item.source.RepositoryID, directory: directory, name: item.file.Name.Name, projectID: projectID}
				groups[key] = group
			}
			group.files = append(group.files, item)
		}
	}
	for _, group := range groups {
		sort.Slice(group.files, func(i, j int) bool { return group.files[i].source.Path < group.files[j].source.Path })
	}
	return groups
}

func (b *graphBuilder) addFile(source Source) {
	id := fileNodeID(source.ID)
	b.addNode(codegraph.Node{ID: id, Kind: codegraph.NodeFile, Name: source.Path, FileID: source.ID, RepositoryID: source.RepositoryID, ProjectIDs: source.ProjectIDs, Evidence: evidence("go/parser", nil)})
}

func (b *graphBuilder) checkGroup(group *packageGroup) {
	packageProjects := projectList(group.projectID)
	packageID := packageNodeID(group.repositoryID, group.directory, group.name, group.projectID)
	b.addNode(codegraph.Node{ID: packageID, Kind: codegraph.NodePackage, Name: group.name, RepositoryID: group.repositoryID, ProjectIDs: packageProjects, Evidence: evidence("go/parser", nil)})
	for _, item := range group.files {
		b.addEdge(codegraph.RelationContains, packageID, fileNodeID(item.source.ID), codegraph.ResolutionExact, packageProjects, evidence("go/parser", nil))
		b.addImports(item)
	}
	if b.truncated || b.ctx.Err() != nil {
		return
	}
	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	files := make([]*ast.File, 0, len(group.files))
	for _, item := range group.files {
		files = append(files, item.file)
	}
	partial := false
	config := types.Config{Importer: sourceOnlyImporter{}, Error: func(error) { partial = true }}
	_, checkErr := config.Check(groupCheckPath(group), b.fset, files, info)
	if checkErr != nil {
		partial = true
	}
	b.complete.CheckedPackages++
	if partial {
		b.complete.PartialPackages++
		b.diagnostic("type_check_partial", "go/types could not completely check the supplied package variant", nil)
	}
	objects := b.addDefinitions(group, info)
	for _, item := range group.files {
		b.addUsesAndCalls(item, info, objects, packageProjects)
		b.addAccessFacts(item, info, objects, packageProjects)
		b.addCallableContainment(item, info, objects, packageProjects)
		if b.truncated || b.ctx.Err() != nil {
			return
		}
	}
	b.addMemberAndTypeRelations(objects, packageProjects)
}

func groupCheckPath(group *packageGroup) string {
	parts := []string{"local", group.repositoryID}
	if group.directory != "" {
		parts = append(parts, group.directory)
	}
	parts = append(parts, group.name)
	if group.projectID != "" {
		parts = append(parts, "project", group.projectID)
	}
	return strings.Join(parts, "/")
}

func (b *graphBuilder) addImports(item parsedSource) {
	for _, spec := range item.file.Imports {
		if b.truncated || b.ctx.Err() != nil {
			return
		}
		rangeValue := b.rangeOf(item.source, spec)
		pathValue := strings.Trim(spec.Path.Value, "\"")
		importID := nodeID("import", item.source.ID, rangeValue.Start, rangeValue.End, pathValue)
		b.addNode(sourceNode(importID, codegraph.NodeImport, pathValue, item.source, rangeValue, "go/parser"))
		b.addEdge(codegraph.RelationContains, fileNodeID(item.source.ID), importID, codegraph.ResolutionExact, item.source.ProjectIDs, evidence("go/parser", &rangeValue))
		b.addEdge(codegraph.RelationImports, importID, "", codegraph.ResolutionUnresolved, item.source.ProjectIDs, evidence("go/parser", &rangeValue))
		b.complete.UnresolvedImports++
		b.diagnostic("unresolved_import", "import was not resolved because only supplied package files are checked", &rangeValue)
	}
}

func (b *graphBuilder) addDefinitions(group *packageGroup, info *types.Info) map[types.Object]string {
	objects := map[types.Object]string{}
	for _, item := range group.files {
		for id, kind := range syntaxDefinitions(item.file) {
			if b.truncated || b.ctx.Err() != nil {
				return objects
			}
			object := info.Defs[id]
			producer := "go/parser"
			if object != nil {
				kind = objectKind(object)
				producer = "go/types"
			}
			symbolID := b.addDefinition(item.source, id, kind, producer)
			if object != nil {
				objects[object] = symbolID
			}
		}
		for id, object := range info.Defs {
			if object == nil || objects[object] != "" || id == nil || id.Name == "_" || !b.belongsTo(item.source, id) {
				continue
			}
			position := b.fset.Position(id.Pos())
			if position.Offset < 0 || position.Offset > len(item.source.Content) {
				continue
			}
			symbolID := b.addDefinition(item.source, id, objectKind(object), "go/types")
			objects[object] = symbolID
		}
	}
	return objects
}

func (b *graphBuilder) addDefinition(source Source, id *ast.Ident, kind, producer string) string {
	rangeValue := b.rangeOf(source, id)
	symbolID := nodeID("symbol", source.ID, rangeValue.Start, rangeValue.End, kind+"\x00"+id.Name)
	definitionID := nodeID("definition", source.ID, rangeValue.Start, rangeValue.End, kind+"\x00"+id.Name)
	b.addNode(sourceNode(symbolID, codegraph.NodeSymbol, id.Name, source, rangeValue, producer))
	b.addNode(sourceNode(definitionID, codegraph.NodeDefinition, id.Name, source, rangeValue, "go/parser"))
	b.addEdge(codegraph.RelationContains, fileNodeID(source.ID), definitionID, codegraph.ResolutionExact, source.ProjectIDs, evidence("go/parser", &rangeValue))
	b.addEdge(codegraph.RelationDeclares, definitionID, symbolID, codegraph.ResolutionExact, source.ProjectIDs, evidence("go/parser", &rangeValue))
	return symbolID
}

func (b *graphBuilder) addUsesAndCalls(item parsedSource, info *types.Info, objects map[types.Object]string, groupProjects []string) {
	ast.Inspect(item.file, func(node ast.Node) bool {
		if b.truncated || b.ctx.Err() != nil {
			return false
		}
		switch value := node.(type) {
		case *ast.Ident:
			if value == item.file.Name || value.Name == "_" || info.Defs[value] != nil {
				return true
			}
			b.addReference(item.source, value, info.Uses[value], objects, groupProjects)
		case *ast.CallExpr:
			b.addCall(item.source, value, callObject(value, info), objects, groupProjects)
		}
		return true
	})
}

func (b *graphBuilder) addReference(source Source, id *ast.Ident, object types.Object, objects map[types.Object]string, groupProjects []string) {
	rangeValue := b.rangeOf(source, id)
	referenceID := nodeID("reference", source.ID, rangeValue.Start, rangeValue.End, id.Name)
	b.addNode(sourceNode(referenceID, codegraph.NodeReference, id.Name, source, rangeValue, "go/parser"))
	b.addEdge(codegraph.RelationContains, fileNodeID(source.ID), referenceID, codegraph.ResolutionExact, source.ProjectIDs, evidence("go/parser", &rangeValue))
	target := objects[object]
	if target != "" {
		b.addEdge(codegraph.RelationRefersTo, referenceID, target, codegraph.ResolutionExact, compatibleProjects(source.ProjectIDs, groupProjects), evidence("go/types", &rangeValue))
		b.complete.ExactReferences++
		return
	}
	b.addEdge(codegraph.RelationRefersTo, referenceID, "", codegraph.ResolutionUnresolved, source.ProjectIDs, evidence("go/types", &rangeValue))
	b.complete.UnresolvedReferences++
}

func (b *graphBuilder) addCall(source Source, call *ast.CallExpr, object types.Object, objects map[types.Object]string, groupProjects []string) {
	rangeValue := b.rangeOf(source, call)
	callID := callSiteID(source, rangeValue)
	b.addNode(sourceNode(callID, codegraph.NodeCallSite, "call", source, rangeValue, "go/parser"))
	b.addEdge(codegraph.RelationContains, fileNodeID(source.ID), callID, codegraph.ResolutionExact, source.ProjectIDs, evidence("go/parser", &rangeValue))
	if _, ok := object.(*types.Func); ok && objects[object] != "" {
		b.addEdge(codegraph.RelationCalls, callID, objects[object], codegraph.ResolutionExact, compatibleProjects(source.ProjectIDs, groupProjects), evidence("go/types", &rangeValue))
		b.complete.ExactCalls++
		return
	}
	b.addEdge(codegraph.RelationCalls, callID, "", codegraph.ResolutionUnresolved, source.ProjectIDs, evidence("go/types", &rangeValue))
	b.complete.UnresolvedCalls++
}

func callObject(call *ast.CallExpr, info *types.Info) types.Object {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return info.Uses[fun]
	case *ast.SelectorExpr:
		if selection := info.Selections[fun]; selection != nil {
			return selection.Obj()
		}
		return info.Uses[fun.Sel]
	}
	return nil
}

func (b *graphBuilder) addMemberAndTypeRelations(objects map[types.Object]string, projects []string) {
	fieldOwners := map[types.Object]types.Object{}
	for owner := range objects {
		typeName, ok := owner.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		if structure, ok := named.Underlying().(*types.Struct); ok {
			for index := 0; index < structure.NumFields(); index++ {
				fieldOwners[structure.Field(index)] = typeName
			}
		}
	}
	for object, symbolID := range objects {
		if b.truncated || b.ctx.Err() != nil {
			return
		}
		owner := memberOwner(object)
		if owner == nil {
			owner = fieldOwners[object]
		}
		if owner != nil && objects[owner] != "" {
			b.addEdge(codegraph.RelationMemberOf, symbolID, objects[owner], codegraph.ResolutionExact, projects, evidence("go/types", nil))
		}
		if typeObject := namedTypeObject(object.Type()); typeObject != nil && typeObject != object && objects[typeObject] != "" {
			b.addEdge(codegraph.RelationHasType, symbolID, objects[typeObject], codegraph.ResolutionExact, projects, evidence("go/types", nil))
		}
	}
}

func memberOwner(object types.Object) types.Object {
	function, ok := object.(*types.Func)
	if !ok {
		return nil
	}
	signature, ok := function.Type().(*types.Signature)
	if !ok || signature.Recv() == nil {
		return nil
	}
	return namedTypeObject(signature.Recv().Type())
}

func namedTypeObject(value types.Type) types.Object {
	switch typed := value.(type) {
	case *types.Named:
		return typed.Obj()
	case *types.Pointer:
		return namedTypeObject(typed.Elem())
	case *types.Alias:
		return namedTypeObject(types.Unalias(typed))
	}
	return nil
}

func syntaxDefinitions(file *ast.File) map[*ast.Ident]string {
	definitions := map[*ast.Ident]string{}
	add := func(id *ast.Ident, kind string) {
		if id != nil && id.Name != "_" {
			definitions[id] = kind
		}
	}
	fields := func(fields *ast.FieldList, kind string) {
		if fields == nil {
			return
		}
		for _, field := range fields.List {
			for _, name := range field.Names {
				add(name, kind)
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncDecl:
			kind := "function"
			if value.Recv != nil {
				kind = "method"
			}
			add(value.Name, kind)
			fields(value.Recv, "receiver")
		case *ast.FuncType:
			fields(value.TypeParams, "type_parameter")
			fields(value.Params, "parameter")
			fields(value.Results, "result")
		case *ast.TypeSpec:
			add(value.Name, "type")
		case *ast.GenDecl:
			for _, spec := range value.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				kind := "variable"
				if value.Tok == token.CONST {
					kind = "constant"
				}
				for _, name := range valueSpec.Names {
					add(name, kind)
				}
			}
		case *ast.StructType:
			fields(value.Fields, "field")
		case *ast.AssignStmt:
			if value.Tok == token.DEFINE {
				for _, expression := range value.Lhs {
					if id, ok := expression.(*ast.Ident); ok {
						add(id, "variable")
					}
				}
			}
		case *ast.RangeStmt:
			if value.Tok == token.DEFINE {
				if id, ok := value.Key.(*ast.Ident); ok {
					add(id, "variable")
				}
				if id, ok := value.Value.(*ast.Ident); ok {
					add(id, "variable")
				}
			}
		}
		return true
	})
	return definitions
}

func objectKind(object types.Object) string {
	switch value := object.(type) {
	case *types.TypeName:
		return "type"
	case *types.Const:
		return "constant"
	case *types.Func:
		if signature, ok := value.Type().(*types.Signature); ok && signature.Recv() != nil {
			return "method"
		}
		return "function"
	case *types.Var:
		if value.IsField() {
			return "field"
		}
		return "variable"
	case *types.PkgName:
		return "import_alias"
	case *types.Label:
		return "label"
	default:
		return "symbol"
	}
}

func (b *graphBuilder) rangeOf(source Source, node ast.Node) codegraph.ByteRange {
	start, end := 0, 0
	if node != nil {
		start = b.fset.Position(node.Pos()).Offset
		end = b.fset.Position(node.End()).Offset
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start > len(source.Content) {
		start = len(source.Content)
	}
	if end > len(source.Content) {
		end = len(source.Content)
	}
	return codegraph.ByteRange{FileID: source.ID, Start: start, End: end}
}

func (b *graphBuilder) belongsTo(source Source, id *ast.Ident) bool {
	position := b.fset.Position(id.Pos())
	return strings.HasSuffix(position.Filename, "@"+source.ID)
}

func (b *graphBuilder) addNode(node codegraph.Node) {
	if b.nodes[node.ID].ID != "" {
		return
	}
	if len(b.nodes) >= MaxNodes {
		b.truncated = true
		return
	}
	node.ProjectIDs = sortedStrings(node.ProjectIDs)
	b.nodes[node.ID] = node
}

func (b *graphBuilder) addEdge(kind codegraph.RelationKind, sourceID, targetID string, resolution codegraph.Resolution, projects []string, evidenceValue codegraph.Evidence) {
	if b.truncated {
		return
	}
	id := edgeID(kind, sourceID, targetID, resolution, evidenceValue.Range, projects)
	if b.edges[id].ID != "" {
		return
	}
	if len(b.edges) >= MaxEdges {
		b.truncated = true
		return
	}
	b.edges[id] = codegraph.Edge{ID: id, Kind: kind, SourceID: sourceID, TargetID: targetID, Resolution: resolution, ProjectIDs: sortedStrings(projects), Evidence: evidenceValue}
}

func (b *graphBuilder) diagnostic(code, message string, rangeValue *codegraph.ByteRange) {
	id := code + "\x00" + message
	if rangeValue != nil {
		id += fmt.Sprintf("\x00%s\x00%d\x00%d", rangeValue.FileID, rangeValue.Start, rangeValue.End)
	}
	b.diagnostics[id] = codegraph.Diagnostic{Code: code, Message: message, Range: rangeValue}
}

func (b *graphBuilder) graph() codegraph.Graph {
	nodes := make([]codegraph.Node, 0, len(b.nodes))
	for _, node := range b.nodes {
		nodes = append(nodes, node)
	}
	edges := make([]codegraph.Edge, 0, len(b.edges))
	for _, edge := range b.edges {
		edges = append(edges, edge)
	}
	diagnostics := make([]codegraph.Diagnostic, 0, len(b.diagnostics))
	for _, diagnostic := range b.diagnostics {
		diagnostics = append(diagnostics, diagnostic)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Code == diagnostics[j].Code {
			return diagnostics[i].Message < diagnostics[j].Message
		}
		return diagnostics[i].Code < diagnostics[j].Code
	})
	b.complete.Truncated = b.truncated
	return codegraph.Graph{Schema: codegraph.SchemaVersion, Nodes: nodes, Edges: edges, Diagnostics: diagnostics, Completeness: b.complete}
}

func sourceNode(id string, kind codegraph.NodeKind, name string, source Source, rangeValue codegraph.ByteRange, producer string) codegraph.Node {
	return codegraph.Node{ID: id, Kind: kind, Name: name, FileID: source.ID, RepositoryID: source.RepositoryID, ProjectIDs: source.ProjectIDs, Range: &rangeValue, Evidence: evidence(producer, &rangeValue)}
}

func evidence(producer string, rangeValue *codegraph.ByteRange) codegraph.Evidence {
	return codegraph.Evidence{Origin: codegraph.OriginAnalyzer, Producer: producer, Range: rangeValue}
}

func fileNodeID(fileID string) string { return nodeID("file", fileID, 0, 0, "") }

func packageNodeID(repositoryID, directory, name, projectID string) string {
	return digest("package", repositoryID, directory, name, projectID)
}

func nodeID(kind, fileID string, start, end int, name string) string {
	return digest(kind, fileID, fmt.Sprint(start), fmt.Sprint(end), name)
}

func edgeID(kind codegraph.RelationKind, sourceID, targetID string, resolution codegraph.Resolution, rangeValue *codegraph.ByteRange, projects []string) string {
	parts := []string{"edge", string(kind), sourceID, targetID, string(resolution), strings.Join(sortedStrings(projects), ",")}
	if rangeValue != nil {
		parts = append(parts, rangeValue.FileID, fmt.Sprint(rangeValue.Start), fmt.Sprint(rangeValue.End))
	}
	return digest(parts...)
}

func digest(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return "cg-" + hex.EncodeToString(hash.Sum(nil)[:16])
}

func projectList(projectID string) []string {
	if projectID == "" {
		return []string{}
	}
	return []string{projectID}
}

func compatibleProjects(fileProjects, groupProjects []string) []string {
	if len(fileProjects) == 0 && len(groupProjects) == 0 {
		return []string{}
	}
	set := map[string]bool{}
	for _, projectID := range fileProjects {
		set[projectID] = true
	}
	result := []string{}
	for _, projectID := range groupProjects {
		if set[projectID] {
			result = append(result, projectID)
		}
	}
	return sortedStrings(result)
}

func sortedStrings(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}

// sourceOnlyImporter rejects imports outside the supplied package group. It
// prevents go/types from reading installed export data or triggering any build.
type sourceOnlyImporter struct{}

// Import always reports an unresolved external import.
func (sourceOnlyImporter) Import(path string) (*types.Package, error) {
	return nil, fmt.Errorf("external import %q is outside supplied graph input", path)
}
