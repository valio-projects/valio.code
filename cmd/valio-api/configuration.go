package main

import (
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/domain"
	"os"
	"strings"
)

type apiConfiguration struct {
	// Address sets the explicit HTTP listening address.
	Address string
	// Token holds the bootstrap credential only in server process memory.
	Token string
	// Workspace defines the only workspace authorized by this bootstrap service.
	Workspace domain.Workspace
	// Origins contains exact browser origins allowed to perform authenticated requests.
	Origins []string
}

func apiConfig() apiConfiguration {
	origins := []string{}
	for _, v := range strings.Split(os.Getenv("VALIO_TRUSTED_ORIGINS"), ",") {
		if strings.TrimSpace(v) != "" {
			origins = append(origins, strings.TrimSpace(v))
		}
	}
	return apiConfiguration{Address: configuration.Env("VALIO_API_ADDRESS", "127.0.0.1:8090"), Token: os.Getenv("VALIO_API_TOKEN"), Workspace: domain.Workspace{ID: domain.WorkspaceID(configuration.Env("VALIO_WORKSPACE_ID", "workspace-main")), Name: "Local workspace"}, Origins: origins}
}
