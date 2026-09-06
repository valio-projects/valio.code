// valio-agent captures local source without running project build tools.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "valio-agent:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("command required: index, doctor, resume, watch")
	}
	switch args[0] {
	case "doctor":
		return runDoctor(out)
	case "index", "watch":
		return runIndex(ctx, args, out)
	case "resume":
		return runResume(ctx, args, out)
	case "help", "-h", "--help":
		fmt.Fprintln(out, "Outbound upload: index, watch, and resume accept --server URL --workspace ID --repository ID. Set VALIO_API_TOKEN in the environment. HTTPS is required except loopback HTTP. Spool data is retained after acceptance; retries use the immutable snapshot identity. Uploads print a server receipt as JSON.")
		_, e := fmt.Fprintln(out, "valio-agent index --root PATH [--output FILE] [--spool DIR]\nvalio-agent doctor\nvalio-agent resume [--spool DIR] [--id SNAPSHOT_ID]\nvalio-agent watch --root PATH [--interval 2s]\nWatch combines fsnotify debounce with periodic full reconciliation; failed uploads remain spooled and retry.\nAll configuration values are excluded, including .env.example defaults. YAML is excluded.\n.valioignore supports *, ?, **, / anchors, trailing / directories, # comments and ! re-inclusion.\nUnsupported character classes/backslash escapes cause capture failure. Last matching rule wins.")
		return e
	default:
		return errors.New("unknown command; use valio-agent help")
	}
}
