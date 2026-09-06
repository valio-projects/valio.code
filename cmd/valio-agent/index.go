package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/valio-projects/valio.code/internal/agent"
	"io"
	"path/filepath"
	"time"
)

func runIndex(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	root := fs.String("root", ".", "Git worktree root")
	output := fs.String("output", "", "sanitized snapshot JSON output file (default stdout)")
	spool := fs.String("spool", "", "offline spool directory (default <root>/.valio/spool)")
	interval := fs.Duration("interval", 2*time.Second, "watch reconciliation interval")
	max := fs.Int64("max-file-bytes", agent.DefaultMaxFileBytes, "file size limit, maximum 2 MiB")
	upload := registerUploadFlags(fs)
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 {
		return errors.New("invalid command flags")
	}
	if *interval < 100*time.Millisecond {
		return errors.New("watch interval must be at least 100ms")
	}
	if *spool == "" {
		*spool = filepath.Join(*root, ".valio", "spool")
	}
	last := ""
	capture := func() error {
		s, e := agent.Capture(ctx, agent.Options{Root: *root, SpoolDir: *spool, MaxFileBytes: *max})
		if e != nil {
			return e
		}
		if s.ID == last {
			return nil
		}
		last = s.ID
		if *output != "" {
			if e := writeOutput(*output, s); e != nil {
				return e
			}
		}
		if upload.Enabled() {
			return upload.Upload(ctx, s, out)
		}
		if *output != "" {
			return nil
		}
		return json.NewEncoder(out).Encode(s)
	}
	if e := capture(); e != nil {
		return e
	}
	if args[0] == "index" {
		return nil
	}
	return reconcileWatch(ctx, *interval, capture)
}
