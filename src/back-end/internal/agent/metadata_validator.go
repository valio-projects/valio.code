package agent

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	gitrepo "github.com/valio-projects/valio.code/internal/git"
)

const (
	maxMachinePathBytes = 32 * 1024
	maxGitRefBytes      = 1024
	maxRepositoryRefs   = 100_000
	maxWorktrees        = 1_024
	maxStatusEntries    = 100_000
)

var gitObjectID = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// validateRepositoryMetadata accepts the bounded syntax emitted by git's
// porcelain commands. It does not attempt to classify arbitrary secret text;
// it only rejects credential-bearing URL patterns in machine paths.
func validateRepositoryMetadata(repository gitrepo.Repository) bool {
	// Repository metadata was absent from early payloads. Preserve that explicit
	// unknown state, but once any metadata is present require the full safe form.
	if repository.Root == "" && repository.GitDir == "" && repository.CommonDir == "" && repository.Head == "" && repository.Branch == "" && len(repository.Refs) == 0 && len(repository.Worktrees) == 0 && len(repository.Status) == 0 && !repository.Dirty {
		return true
	}
	if !validMachinePath(repository.Root) || !validMachinePath(repository.GitDir) || !validMachinePath(repository.CommonDir) ||
		!validObjectIDOrEmpty(repository.Head) || len(repository.Refs) > maxRepositoryRefs || len(repository.Worktrees) > maxWorktrees || len(repository.Status) > maxStatusEntries ||
		repository.Dirty != (len(repository.Status) > 0) {
		return false
	}
	if !sort.StringsAreSorted(repository.Refs) || !uniqueStrings(repository.Refs) {
		return false
	}
	for _, ref := range repository.Refs {
		if !validRef(ref) {
			return false
		}
	}
	for i, worktree := range repository.Worktrees {
		if !validMachinePath(worktree.Path) || !validObjectIDOrEmpty(worktree.Head) || !validWorktreeBranch(worktree.Branch) {
			return false
		}
		if i > 0 && repository.Worktrees[i-1].Path >= worktree.Path {
			return false
		}
	}
	for i, status := range repository.Status {
		if !validStatus(status) {
			return false
		}
		if i > 0 && repository.Status[i-1].Path > status.Path {
			return false
		}
	}
	return validBranch(repository.Branch)
}

func validMachinePath(value string) bool {
	if value == "" || len(value) > maxMachinePathBytes || !utf8.ValidString(value) || containsControl(value) || credentialURL.MatchString(value) {
		return false
	}
	// Git records absolute filesystem paths. Accept Windows drive and UNC paths
	// on every host so a snapshot remains portable to a validating service.
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, `\\`) || (len(value) >= 3 && isASCIIAlpha(value[0]) && value[1] == ':' && (value[2] == '/' || value[2] == '\\')) {
		return true
	}
	return false
}

func validObjectIDOrEmpty(value string) bool { return value == "" || gitObjectID.MatchString(value) }

func validBranch(value string) bool {
	if value == "" {
		return true // An unborn HEAD has no symbolic branch value to persist.
	}
	if strings.HasPrefix(value, "refs/") {
		return validRef(value)
	}
	return validRef("refs/heads/" + value)
}

func validWorktreeBranch(value string) bool {
	return value == "" || (strings.HasPrefix(value, "refs/") && validRef(value))
}

func validRef(value string) bool {
	if len(value) == 0 || len(value) > maxGitRefBytes || !utf8.ValidString(value) || containsControl(value) || !strings.HasPrefix(value, "refs/") || strings.HasSuffix(value, "/") || strings.Contains(value, "..") || strings.Contains(value, "@{") || strings.ContainsAny(value, "\\ ~^:?*[") {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || component == "." || strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}

func validStatus(status gitrepo.Status) bool {
	return len(status.Index) == 1 && len(status.Worktree) == 1 && strings.ContainsRune(" MADRCU?!T", rune(status.Index[0])) && strings.ContainsRune(" MADRCU?!T", rune(status.Worktree[0])) && safeRelativePath(status.Path) && (status.OriginalPath == "" || safeRelativePath(status.OriginalPath))
}

func uniqueStrings(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1] == values[i] {
			return false
		}
	}
	return true
}

func containsControl(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

func isASCIIAlpha(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}
