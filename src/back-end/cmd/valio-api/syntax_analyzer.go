package main

import (
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/infrastructure/syntax"
	"os"
)

// configureSyntax enables only an operator-selected parser helper. A missing
// configured executable is a startup error, never silently reported as ready.
func configureSyntax(service *snapshots.Service) error {
	script := os.Getenv("VALIO_SYNTAX_HELPER")
	if script == "" {
		return nil
	}
	node := os.Getenv("VALIO_NODE_BINARY")
	if node == "" {
		node = "node"
	}
	processor, err := syntax.New(node, script)
	if err != nil {
		return err
	}
	service.Syntax = processor
	return nil
}
