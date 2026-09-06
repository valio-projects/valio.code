package main

import (
	"context"
	"fmt"
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/scheduling"
	"github.com/valio-projects/valio.code/internal/storage/surreal"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := configuration.Database()
	if err != nil {
		return err
	}
	db, err := surreal.New(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if len(os.Args) == 2 && os.Args[1] == "health" {
		return db.Ping(ctx)
	}
	if err = db.Ping(ctx); err != nil {
		return err
	}
	host, _ := os.Hostname()
	q := scheduling.Queue{DB: db}
	// Supported job handlers are registered explicitly. Unknown work is failed
	// as unsupported rather than acknowledged as a successful analysis.
	err = q.Run(ctx, fmt.Sprintf("%s-%d", host, os.Getpid()), map[string]scheduling.Handler{"health": func(ctx context.Context, _ scheduling.Job) error { return db.Ping(ctx) }})
	if ctx.Err() != nil {
		return nil
	}
	return err
}
