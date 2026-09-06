// valio-mcp is a native stdio bridge; indexing outlives its connection.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/valio-projects/valio.code/internal/transport/mcp"
	"os"
	"os/signal"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: valio-mcp serve --server URL (VALIO_API_TOKEN required)")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	server := flags.String("server", "http://127.0.0.1:8080", "valio API origin")
	_ = flags.Parse(os.Args[2:])
	bridge, err := mcp.NewBridge(*server, os.Getenv("VALIO_API_TOKEN"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err = bridge.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
