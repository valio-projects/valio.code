package analysis

type Options struct {
	// CheckTypes runs the real Go type checker on this one file. Other files in
	// a package and module-specific dependencies are not supplied implicitly.
	CheckTypes  bool
	PackagePath string
}
