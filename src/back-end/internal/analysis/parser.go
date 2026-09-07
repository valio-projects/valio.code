package analysis

// Parser owns the analysis policy. The zero value extracts syntax evidence;
// Options.CheckTypes enables actual single-file Go type checking.
type Parser struct {
	Options  Options
	Syntax   SyntaxProvider
	Compiler CompilerProvider
}

func Analyze(path, content string) Report { return (Parser{}).Analyze(path, content) }
func AnalyzeWithOptions(path, content string, options Options) Report {
	return (Parser{Options: options}).Analyze(path, content)
}
