package domain

import "fmt"

type EvidenceOrigin string

const (
	OriginAnalyzer EvidenceOrigin = "analyzer"
	OriginImport   EvidenceOrigin = "import"
	OriginHuman    EvidenceOrigin = "human"
)

type EvidenceResolution string

const (
	ResolutionUnresolved EvidenceResolution = "unresolved"
	ResolutionConfirmed  EvidenceResolution = "confirmed"
	ResolutionRejected   EvidenceResolution = "rejected"
	ResolutionFixed      EvidenceResolution = "fixed"
)

type EvidenceAssertion string

const (
	AssertionObserved EvidenceAssertion = "observed"
	AssertionInferred EvidenceAssertion = "inferred"
	AssertionReported EvidenceAssertion = "reported"
)

type PositionEncoding string

const (
	EncodingUTF8  PositionEncoding = "utf-8"
	EncodingUTF16 PositionEncoding = "utf-16"
)

// Position is zero-based. Character counts code units in SourceRange.Encoding.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// SourceRange is half-open: Start is included and End is excluded.
type SourceRange struct {
	RepositoryID RepositoryID     `json:"repositoryId"`
	Path         string           `json:"path"`
	Encoding     PositionEncoding `json:"encoding"`
	Start        Position         `json:"start"`
	End          Position         `json:"end"`
}

func (r SourceRange) Validate() error {
	if r.RepositoryID == "" {
		return fmt.Errorf("source range requires repository identity")
	}
	if err := ValidateRelativePath(r.Path, false); err != nil {
		return err
	}
	if r.Encoding != EncodingUTF8 && r.Encoding != EncodingUTF16 {
		return fmt.Errorf("unsupported position encoding %q", r.Encoding)
	}
	if r.Start.Line < 0 || r.Start.Character < 0 || r.End.Line < 0 || r.End.Character < 0 || r.End.Line < r.Start.Line || (r.End.Line == r.Start.Line && r.End.Character < r.Start.Character) {
		return fmt.Errorf("invalid half-open source range")
	}
	return nil
}

// Origin, resolution, and assertion are orthogonal; e.g. a human report may be
// rejected without altering its origin or what was originally asserted.
type Evidence struct {
	ID         string             `json:"id"`
	ViewID     ViewID             `json:"viewId"`
	Origin     EvidenceOrigin     `json:"origin"`
	Resolution EvidenceResolution `json:"resolution"`
	Assertion  EvidenceAssertion  `json:"assertion"`
	Range      SourceRange        `json:"range"`
	Summary    string             `json:"summary"`
}

func (e Evidence) Validate() error {
	if e.ID == "" || e.ViewID == "" {
		return fmt.Errorf("evidence requires identity and view")
	}
	if e.Origin != OriginAnalyzer && e.Origin != OriginImport && e.Origin != OriginHuman {
		return fmt.Errorf("invalid origin")
	}
	if e.Resolution != ResolutionUnresolved && e.Resolution != ResolutionConfirmed && e.Resolution != ResolutionRejected && e.Resolution != ResolutionFixed {
		return fmt.Errorf("invalid resolution")
	}
	if e.Assertion != AssertionObserved && e.Assertion != AssertionInferred && e.Assertion != AssertionReported {
		return fmt.Errorf("invalid assertion")
	}
	return e.Range.Validate()
}
