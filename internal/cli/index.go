package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/skill"
)

func runIndex(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var roots stringList
	flags.Var(&roots, "root", "skill root as [namespace=]path; repeatable")
	database := flags.String("db", registry.DefaultPath(), "registry database")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if len(roots) == 0 {
		fmt.Fprintln(stderr, "at least one --root is required")
		return ExitInvalidInput
	}
	parsedRoots := make([]skill.Root, 0, len(roots))
	for _, value := range roots {
		namespace, path := "", value
		if before, after, found := strings.Cut(value, "="); found {
			namespace, path = before, after
		}
		if strings.TrimSpace(path) == "" {
			fmt.Fprintf(stderr, "invalid root %q\n", value)
			return ExitInvalidInput
		}
		parsedRoots = append(parsedRoots, skill.Root{Path: path, Namespace: namespace})
	}
	store, err := registry.Open(*database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	defer store.Close()
	generation, err := registry.Index(context.Background(), store, parsedRoots)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	if *jsonOutput {
		if err := writeJSON(stdout, generation); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitRegistryUnhealthy
		}
	} else {
		fmt.Fprintf(stdout, "indexed %d skills (%d diagnostics) %s\n", generation.SkillCount, generation.DiagnosticCount, generation.Fingerprint)
	}
	return ExitOK
}
