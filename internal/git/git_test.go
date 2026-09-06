package git

import (
	"testing"
)

func TestParseRenameStatus(t *testing.T) {
	s := parseStatus([]byte("R  new name.go\x00old name.go\x00?? untracked.go\x00"))
	if len(s) != 2 || s[0].Path != "new name.go" || s[0].OriginalPath != "old name.go" || s[1].Index != "?" {
		t.Fatalf("bad rename status: %#v", s)
	}
}
func TestParseWorktrees(t *testing.T) {
	s := parseWorktrees([]byte("worktree /repo\x00HEAD abc\x00branch refs/heads/main\x00\x00worktree /other\x00HEAD def\x00detached\x00\x00"))
	if len(s) != 2 || s[0].Path != "/other" || !s[0].Detached || s[1].Branch != "refs/heads/main" {
		t.Fatalf("bad worktrees: %#v", s)
	}
}
