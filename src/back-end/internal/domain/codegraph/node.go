package codegraph

// NodeKind classifies one graph node without implying semantic completeness.
type NodeKind string

const (
	// NodeFile represents one supplied source file.
	NodeFile NodeKind = "file"
	// NodePackage represents one checked package variant in one repository/directory/project context.
	NodePackage NodeKind = "package"
	// NodeImport represents one written import declaration.
	NodeImport NodeKind = "import"
	// NodeExternalPackage represents an import path not resolved from the supplied input.
	NodeExternalPackage NodeKind = "external_package"
	// NodeDefinition represents one source declaration span.
	NodeDefinition NodeKind = "definition"
	// NodeSymbol represents one locally declared Go object.
	NodeSymbol NodeKind = "symbol"
	// NodeReference represents one identifier use.
	NodeReference NodeKind = "reference"
	// NodeCallSite represents one written call expression.
	NodeCallSite NodeKind = "call_site"
)

// Node is one version-independent local graph identity.
type Node struct {
	// ID is stable for a supplied file identity, range, kind, and semantic name.
	ID string `json:"id"`
	// Kind classifies this node.
	Kind NodeKind `json:"kind"`
	// Name is the written or compiler-reported name, when one exists.
	Name string `json:"name,omitempty"`
	// FileID identifies the local input file when this node is source-backed.
	FileID string `json:"fileId,omitempty"`
	// RepositoryID scopes the node to one repository input.
	RepositoryID string `json:"repositoryId"`
	// ProjectIDs records every project context contributed by the source file or package variant.
	ProjectIDs []string `json:"projectIds"`
	// Range is present for source-backed nodes and counts UTF-8 bytes.
	Range *ByteRange `json:"range,omitempty"`
	// Evidence identifies the observation that produced this node.
	Evidence Evidence `json:"evidence"`
}
