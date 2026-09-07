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
	origins := []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	if configured := os.Getenv("VALIO_TRUSTED_ORIGINS"); configured != "" {
		origins = nil
		for _, v := range strings.Split(configured, ",") {
			if strings.TrimSpace(v) != "" {
				origins = append(origins, strings.TrimSpace(v))
			}
		}
	}
	return apiConfiguration{Address: configuration.Env("VALIO_API_ADDRESS", configuration.DefaultAPIAddress), Token: configuration.ApplicationToken(), Workspace: domain.Workspace{ID: domain.WorkspaceID(configuration.Env("VALIO_WORKSPACE_ID", configuration.DefaultWorkspaceID)), Name: "Local workspace"}, Origins: origins}
}
