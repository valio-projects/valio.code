package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"

	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/configuration"
)

type uploadFlags struct{ server, workspace, repository *string }

func registerUploadFlags(fs *flag.FlagSet) uploadFlags {
	return uploadFlags{fs.String("server", "", "optional ingestion server URL"), fs.String("workspace", "workspace-main", "workspace ID"), fs.String("repository", "repo-main", "repository ID")}
}
func (f uploadFlags) Enabled() bool { return *f.server != "" }

// Validate rejects invalid destinations before capture or spool side effects.
func (f uploadFlags) Validate() error {
	if !f.Enabled() {
		return nil
	}
	return (agent.UploadOptions{Endpoint: *f.server, Token: configuration.ApplicationToken(), WorkspaceID: *f.workspace, RepositoryID: *f.repository}).Validate()
}
func (f uploadFlags) Upload(ctx context.Context, s agent.Snapshot, out io.Writer) error {
	client := agent.UploadClient{Options: agent.UploadOptions{Endpoint: *f.server, Token: configuration.ApplicationToken(), WorkspaceID: *f.workspace, RepositoryID: *f.repository}}
	result, e := client.Upload(ctx, s)
	if e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(result)
}
