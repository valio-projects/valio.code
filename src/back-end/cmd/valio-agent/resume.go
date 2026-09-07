package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/valio-projects/valio.code/internal/agent"
	"io"
)

func runResume(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("resume", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dir := fs.String("spool", ".valio/spool", "spool directory")
	id := fs.String("id", "", "snapshot ID to emit (default validate/list pending IDs)")
	upload := registerUploadFlags(fs)
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 {
		return errors.New("invalid resume flags")
	}
	if *id != "" {
		s, e := agent.Load(*dir, *id)
		if e != nil {
			return e
		}
		if upload.Enabled() {
			return upload.Upload(ctx, s, out)
		}
		return json.NewEncoder(out).Encode(s)
	}
	ids, e := agent.Pending(*dir)
	if e != nil {
		return e
	}
	if upload.Enabled() {
		for _, pendingID := range ids {
			s, e := agent.Load(*dir, pendingID)
			if e != nil {
				return e
			}
			if e = upload.Upload(ctx, s, out); e != nil {
				return e
			}
		}
		return nil
	}
	return json.NewEncoder(out).Encode(struct {
		IDs    []string `json:"snapshotIds"`
		Status string   `json:"status"`
	}{ids, "validated-local; not uploaded"})
}
