package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	profileconfig "github.com/valio-projects/valio.code/internal/configuration/ai"
)

type localProbe func(context.Context) []profileconfig.LocalProbe

func runAIDoctor(args []string) error {
	return runAIDoctorWith(args, profileconfig.ProbeLocal, os.Stdout)
}

func runAIDoctorWith(args []string, probe localProbe, output io.Writer) error {
	flags := flag.NewFlagSet("ai doctor", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	configPath := flags.String("config", "", "path to AI profile JSON")
	profileID := flags.String("profile", "", "profile ID to validate with one synthetic embedding")
	list := flags.Bool("list", false, "list sanitized configured profiles")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *configPath == "" {
		return errors.New("usage: valio-admin ai doctor --config PATH [--profile ID] [--list]")
	}
	document, err := profileconfig.Load(*configPath)
	if err != nil {
		return err
	}
	if *list {
		writeProfiles(output, document)
		if *profileID == "" {
			return nil
		}
	}
	if *profileID == "" {
		return errors.New("ai doctor requires --profile unless --list is selected")
	}
	profile, embedder, err := document.Resolve(*profileID)
	if err != nil {
		return err
	}
	probeContext, cancelProbe := context.WithTimeout(context.Background(), 2*time.Second)
	for _, result := range probe(probeContext) {
		fmt.Fprintf(output, "provider=%s status=%s\n", result.Provider, result.Status)
	}
	cancelProbe()
	embeddingTimeout, err := profileTimeout(document, *profileID)
	if err != nil {
		return err
	}
	embeddingContext, cancelEmbedding := context.WithTimeout(context.Background(), embeddingTimeout)
	defer cancelEmbedding()
	vector, err := embedder.Embed(embeddingContext, profile, "valio doctor synthetic embedding")
	if err != nil {
		return errors.New("AI profile embedding probe failed")
	}
	if len(vector) != profile.Dimension {
		return errors.New("AI profile embedding dimension mismatch")
	}
	fmt.Fprintf(output, "profile=%s status=available dimension=%d\n", profile.ID, profile.Dimension)
	return nil
}

func profileTimeout(document profileconfig.Document, profileID string) (time.Duration, error) {
	for _, profile := range document.Profiles {
		if profile.ID != profileID {
			continue
		}
		timeout, err := time.ParseDuration(profile.Timeout)
		if err != nil || timeout <= 0 || timeout > 120*time.Second {
			return 0, errors.New("invalid AI profile timeout")
		}
		return timeout, nil
	}
	return 0, errors.New("AI profile not found")
}

func writeProfiles(output io.Writer, document profileconfig.Document) {
	profiles := append([]profileconfig.Profile{}, document.Profiles...)
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	for _, profile := range profiles {
		fmt.Fprintf(output, "profile=%s provider=%s model=%s dimension=%d kind=%s\n", profile.ID, profile.ProviderKind, profile.Model, profile.Dimension, profile.Kind)
	}
}
