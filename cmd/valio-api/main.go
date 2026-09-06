package main

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/infrastructure/telemetry"
	"go.uber.org/fx"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "health" {
		if err := apiHealth(); err != nil {
			fmt.Fprintln(os.Stderr, "API readiness check failed")
			os.Exit(1)
		}
		return
	}
	app := fx.New(telemetry.Module, apiModule())
	if err := app.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "API configuration failed:", err)
		os.Exit(1)
	}
	app.Run()
}
