package snapshots

func supportedSyntax(language string) bool {
	switch language {
	case "c", "cpp", "csharp", "javascript", "typescript", "java":
		return true
	}
	return false
}
