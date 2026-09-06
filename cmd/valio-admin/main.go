package main

import (
	"context"
	"fmt"
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "migrate" && os.Args[1] != "health") {
		return fmt.Errorf("usage: valio-admin migrate|health")
	}
	cfg, err := configuration.Database()
	if err != nil {
		return err
	}
	db, err := surreal.New(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if os.Args[1] == "health" {
		return db.Ping(ctx)
	}
	if err = db.Migrate(ctx); err != nil {
		return err
	}
	for _, role := range []struct{ name, env string }{{"valio_api", "VALIO_API_DB_PASSWORD"}, {"valio_worker", "VALIO_WORKER_DB_PASSWORD"}} {
		if password := os.Getenv(role.env); password != "" {
			if err = db.ProvisionUser(ctx, role.name, password, "EDITOR"); err != nil {
				return err
			}
		}
	}
	fmt.Println("Database schema migrated:", db.Database())
	return nil
}
