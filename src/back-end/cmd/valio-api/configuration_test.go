package main

import (
	"reflect"
	"testing"

	"github.com/valio-projects/valio.code/internal/configuration"
)

func TestAPIConfigUsesLocalDefaults(t *testing.T) {
	t.Setenv("VALIO_API_ADDRESS", "")
	t.Setenv("VALIO_API_TOKEN", "")
	t.Setenv("VALIO_WORKSPACE_ID", "")
	t.Setenv("VALIO_TRUSTED_ORIGINS", "")

	config := apiConfig()
	if config.Address != configuration.DefaultAPIAddress || config.Token != configuration.DefaultAPIToken || string(config.Workspace.ID) != configuration.DefaultWorkspaceID {
		t.Fatalf("unexpected local defaults: %#v", config)
	}
	wantOrigins := []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	if !reflect.DeepEqual(config.Origins, wantOrigins) {
		t.Fatalf("trusted origins = %#v, want %#v", config.Origins, wantOrigins)
	}
}

func TestAPIConfigAppliesEnvironmentOverrides(t *testing.T) {
	t.Setenv("VALIO_API_ADDRESS", "127.0.0.1:8099")
	t.Setenv("VALIO_API_TOKEN", "configured-token")
	t.Setenv("VALIO_WORKSPACE_ID", "configured-workspace")
	t.Setenv("VALIO_TRUSTED_ORIGINS", " https://ui.example ,https://other.example ")

	config := apiConfig()
	if config.Address != "127.0.0.1:8099" || config.Token != "configured-token" || string(config.Workspace.ID) != "configured-workspace" {
		t.Fatalf("environment overrides were not applied: %#v", config)
	}
	wantOrigins := []string{"https://ui.example", "https://other.example"}
	if !reflect.DeepEqual(config.Origins, wantOrigins) {
		t.Fatalf("trusted origins = %#v, want %#v", config.Origins, wantOrigins)
	}
}
