// valio-worker composes repositories and processors with Uber Fx.
package main

import (
	"context"
	"fmt"
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"github.com/valio-projects/valio.code/internal/infrastructure/telemetry"
	"go.uber.org/fx"
	"os"
	"time"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "health" {
		if err := health(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fx.New(telemetry.Module, workerModule).Run()
}
func health() error {
	cfg, err := configuration.Database()
	if err != nil {
		return err
	}
	db, err := surreal.New(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return db.Ping(ctx)
}
