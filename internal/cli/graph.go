package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/mistakeknot/intermesh/internal/registry"
)

func runGraph(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("graph", flag.ContinueOnError)
	flags.SetOutput(stderr)
	database := flags.String("db", registry.DefaultPath(), "registry database")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if len(flags.Args()) != 1 {
		fmt.Fprintln(stderr, "exactly one skill ID is required")
		return ExitInvalidInput
	}
	store, err := registry.Open(*database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	defer store.Close()
	generation, err := store.Current(context.Background())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	for _, item := range generation.Skills {
		if item.ID != flags.Args()[0] {
			continue
		}
		if *jsonOutput {
			_ = writeJSON(stdout, item)
		} else {
			fmt.Fprintf(stdout, "%s\n  requires: %v\n  composes_with: %v\n  conflicts_with: %v\n  supersedes: %v\n", item.ID, item.Manifest.Requires, item.Manifest.ComposesWith, item.Manifest.ConflictsWith, item.Manifest.Supersedes)
		}
		return ExitOK
	}
	fmt.Fprintf(stderr, "skill %q not found\n", flags.Args()[0])
	return ExitGraphUnresolved
}
