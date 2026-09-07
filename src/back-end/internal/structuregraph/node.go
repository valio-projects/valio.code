package structuregraph

// NodeKind classifies a syntactic graph observation.
type NodeKind string

const (
	// NodeFile represents one supplied source file.
	NodeFile NodeKind = "file"
	// NodeDeclaration represents one declaration reported by the syntax helper.
	NodeDeclaration NodeKind = "declaration"
	// NodeParameter represents one declared parameter of a reported declaration.
	NodeParameter NodeKind = "parameter"
	// NodeImport represents one written import observation.
	NodeImport NodeKind = "import"
	// NodeReference represents one non-call identifier reference observation.
	NodeReference NodeKind = "reference"
	// NodeCall represents one syntactic call observation whose target is unknown.
	NodeCall NodeKind = "call"
)

// Node stores one source-backed syntactic fact. Fields copied from a helper
// report describe spelling and annotations only; they do not establish semantic resolution.
type Node struct {
	// ID is deterministic from the supplied file ID, report identifier, and range.
	ID string `json:"id"`
	// Kind classifies this observed fact.
	Kind NodeKind `json:"kind"`
	// Name is the helper-reported written name, if present.
	Name string `json:"name,omitempty"`
	// DeclarationKind is the helper's declaration category for NodeDeclaration.
	DeclarationKind string `json:"declarationKind,omitempty"`
	// FileID identifies the supplied source file that owns this node.
	FileID string `json:"fileId"`
	// RepositoryID scopes the node to one repository.
	RepositoryID string `json:"repositoryId"`
	// ProjectIDs records every project context for the owning source.
	ProjectIDs []string `json:"projectIds"`
	// Range identifies this fact's UTF-8 source span.
	Range ByteRange `json:"range"`
	// Type is the helper-reported written type string, when available.
	Type *string `json:"type,omitempty"`
	// UnderlyingType is the helper-reported enum backing type, when available.
	UnderlyingType *string `json:"underlyingType,omitempty"`
	// EnumValue is the helper-reported enum member value, when available.
	EnumValue *string `json:"enumValue,omitempty"`
	// Visibility is the helper-reported access spelling, or unknown when absent.
	Visibility string `json:"visibility,omitempty"`
	// Modifiers are direct syntactic modifiers reported by the helper.
	Modifiers []string `json:"modifiers,omitempty"`
	// Attributes are direct syntactic attributes or annotations reported by the helper.
	Attributes []string `json:"attributes,omitempty"`
	// Evidence identifies the helper observation supporting this node.
	Evidence Evidence `json:"evidence"`
}
