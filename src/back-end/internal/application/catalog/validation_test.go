package catalog

import (
	"context"
	"errors"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"testing"
)

func TestRepositoryURLsRejectedBeforePersistence(t *testing.T) {
	service := Service{Workspace: domain.Workspace{ID: "w"}}
	for _, remote := range []string{"https://example.com:70000/repo", "ssh://bad_host/repo", "https://example.com/%2e%2e/repo", "https://user:secret@example.com/repo", "https://example.com?"} {
		_, err := service.Register(context.Background(), domain.Repository{ID: "repo", WorkspaceID: "w", RemoteURL: remote})
		if !errors.Is(err, fault.ErrInvalid) {
			t.Fatalf("expected invalid before nil store: %v", err)
		}
	}
}
