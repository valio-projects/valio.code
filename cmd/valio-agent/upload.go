package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/valio-projects/valio.code/internal/agent"
)

type uploadFlags struct{ server, workspace, repository *string }

func registerUploadFlags(fs *flag.FlagSet) uploadFlags {
	return uploadFlags{fs.String("server", "", "optional ingestion server URL"), fs.String("workspace", "workspace-main", "workspace ID"), fs.String("repository", "repo-main", "repository ID")}
}
func (f uploadFlags) Enabled() bool { return *f.server != "" }
func (f uploadFlags) Upload(ctx context.Context, s agent.Snapshot, out io.Writer) error {
	client := agent.UploadClient{Options: agent.UploadOptions{Endpoint: *f.server, Token: os.Getenv("VALIO_API_TOKEN"), WorkspaceID: *f.workspace, RepositoryID: *f.repository}}
	result, e := client.Upload(ctx, s)
	if e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(result)
}
