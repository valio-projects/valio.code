package main

import (
	"go.uber.org/fx"
	"testing"
)

func TestDependencyGraph(t *testing.T) {
	if e := fx.ValidateApp(apiModule()); e != nil {
		t.Fatal(e)
	}
}
