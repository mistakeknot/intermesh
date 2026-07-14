package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/mistakeknot/intermesh/internal/graph"
	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
)

func runResolve(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("resolve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	database := flags.String("db", registry.DefaultPath(), "registry database")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if len(flags.Args()) == 0 {
		fmt.Fprintln(stderr, "at least one skill ID is required")
		return ExitInvalidInput
	}
	store, err := registry.Open(*database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	defer store.Close()
	generation, err := store.Current(context.Background())
	if err != nil || generation.Fingerprint == "" {
		fmt.Fprintln(stderr, "registry is empty or unreadable")
		return ExitRegistryUnhealthy
	}
	input := route.Result{Version: 1, RegistryFingerprint: generation.Fingerprint, Candidates: []route.Candidate{}, Warnings: []string{}}
	for _, id := range flags.Args() {
		input.Candidates = append(input.Candidates, route.Candidate{ID: id, SelectedBy: "explicit", Reasons: []string{}, RequiredBy: []string{}})
	}
	result, err := graph.Resolve(input, generation)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitGraphUnresolved
	}
	if *jsonOutput {
		_ = writeJSON(stdout, result)
	} else {
		for _, candidate := range result.Candidates {
			fmt.Fprintln(stdout, candidate.ID)
		}
	}
	return ExitOK
}
