package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mistakeknot/intermesh/internal/graph"
	"github.com/mistakeknot/intermesh/internal/receipt"
	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
)

func runRoute(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("route", flag.ContinueOnError)
	flags.SetOutput(stderr)
	query := flags.String("query", "", "request to route")
	host := flags.String("host", "", "agent host")
	database := flags.String("db", registry.DefaultPath(), "registry database")
	limit := flags.Int("limit", route.DefaultLimit, "maximum ranked candidates")
	environment := flags.String("environment", "", "environment gate")
	cwd := flags.String("cwd", "", "working directory")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	noReceipt := flags.Bool("no-receipt", false, "do not append a route receipt")
	receiptPath := flags.String("receipt", "", "receipt JSONL path")
	receiptQuery := flags.String("receipt-query", "hash", "hash or plain")
	var extensions stringList
	flags.Var(&extensions, "extension", "active file extension; repeatable")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if *query == "" {
		fmt.Fprintln(stderr, "--query is required")
		return ExitInvalidInput
	}
	if *limit < 1 {
		fmt.Fprintln(stderr, "--limit must be positive")
		return ExitInvalidInput
	}
	if *receiptQuery != "hash" && *receiptQuery != "plain" {
		fmt.Fprintln(stderr, "--receipt-query must be hash or plain")
		return ExitInvalidInput
	}
	if *cwd == "" {
		*cwd, _ = os.Getwd()
	}
	store, err := registry.Open(*database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitRegistryUnhealthy
	}
	defer store.Close()
	generation, err := store.Current(context.Background())
	if err != nil || generation.Fingerprint == "" {
		fmt.Fprintln(stderr, "registry is empty or unreadable; run intermesh index")
		return ExitRegistryUnhealthy
	}
	started := time.Now()
	result := route.Rank(route.Request{
		Query: *query, Host: *host, CWD: *cwd, Limit: *limit,
		Environment: *environment, Extensions: extensions,
	}, generation)
	result, err = graph.Resolve(result, generation)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitGraphUnresolved
	}
	latency := time.Since(started)
	if !*noReceipt {
		writer := receipt.Writer{Path: *receiptPath, IncludeQuery: *receiptQuery == "plain"}
		if err := writer.Append(context.Background(), result, receipt.Metadata{CWD: *cwd, Limit: *limit, Latency: latency}); err != nil {
			fmt.Fprintf(stderr, "receipt: %v\n", err)
		}
	}
	if *jsonOutput {
		if err := writeJSON(stdout, result); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitRegistryUnhealthy
		}
	} else {
		for _, candidate := range result.Candidates {
			fmt.Fprintf(stdout, "%s\t%.3f\t%s\n", candidate.ID, candidate.Score, candidate.SkillMD)
		}
		for _, warning := range result.Warnings {
			fmt.Fprintf(stderr, "warning: %s\n", warning)
		}
	}
	return ExitOK
}
