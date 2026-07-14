package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
)

func runSearch(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("search", flag.ContinueOnError)
	flags.SetOutput(stderr)
	database := flags.String("db", registry.DefaultPath(), "registry database")
	limit := flags.Int("limit", 20, "maximum results")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	query := strings.Join(flags.Args(), " ")
	if query == "" {
		fmt.Fprintln(stderr, "search query is required")
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
	result := route.Rank(route.Request{Query: query, Limit: *limit}, generation)
	if *jsonOutput {
		if err := writeJSON(stdout, result); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitRegistryUnhealthy
		}
	} else {
		for _, candidate := range result.Candidates {
			fmt.Fprintf(stdout, "%s\t%.3f\t%s\n", candidate.ID, candidate.Score, candidate.SkillMD)
		}
	}
	return ExitOK
}
