package analysis

type Diagnostic struct {
	Message string `json:"message"`
	Offset  int    `json:"offset"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}
