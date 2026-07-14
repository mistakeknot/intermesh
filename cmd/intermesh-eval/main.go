package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	intermesheval "github.com/mistakeknot/intermesh/internal/eval"
	"github.com/mistakeknot/intermesh/internal/graph"
	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
)

type latencyReport struct {
	Samples int   `json:"samples"`
	P50     int64 `json:"p50_micros"`
	P95     int64 `json:"p95_micros"`
}

type report struct {
	RegistryFingerprint string                           `json:"registry_fingerprint"`
	Metrics             intermesheval.Metrics            `json:"metrics"`
	ByKind              map[string]intermesheval.Metrics `json:"by_kind"`
	Latency             latencyReport                    `json:"warm_latency"`
	Predictions         map[string][]string              `json:"predictions"`
}

func main() {
	database := flag.String("db", registry.DefaultPath(), "registry database")
	casesPath := flag.String("cases", "", "evaluation cases JSONL")
	host := flag.String("host", "codex", "default host")
	limit := flag.Int("limit", 5, "candidate limit")
	runs := flag.Int("runs", 20, "measured runs per case")
	warmups := flag.Int("warmups", 2, "warmup runs per case")
	flag.Parse()
	if *casesPath == "" || *limit < 1 || *runs < 1 || *warmups < 0 {
		fmt.Fprintln(os.Stderr, "--cases, positive --limit/--runs, and nonnegative --warmups are required")
		os.Exit(2)
	}
	file, err := os.Open(*casesPath)
	if err != nil {
		fatal(err)
	}
	defer file.Close()
	cases, err := intermesheval.LoadCases(file)
	if err != nil {
		fatal(err)
	}
	store, err := registry.Open(*database)
	if err != nil {
		fatal(err)
	}
	defer store.Close()
	generation, err := store.Current(context.Background())
	if err != nil || generation.Fingerprint == "" {
		fatal(fmt.Errorf("registry is empty or unreadable"))
	}

	predictions := make(map[string][]string, len(cases))
	timings := make([]int64, 0, len(cases)*(*runs))
	for _, item := range cases {
		caseHost := item.Host
		if caseHost == "" {
			caseHost = *host
		}
		request := route.Request{Query: item.Query, Host: caseHost, Limit: *limit, Environment: item.Environment, Extensions: item.Extensions}
		for index := 0; index < *warmups; index++ {
			if _, err := evaluate(request, generation); err != nil {
				fatal(fmt.Errorf("case %s: %w", item.ID, err))
			}
		}
		var selected []string
		for index := 0; index < *runs; index++ {
			started := time.Now()
			result, err := evaluate(request, generation)
			if err != nil {
				fatal(fmt.Errorf("case %s: %w", item.ID, err))
			}
			timings = append(timings, time.Since(started).Microseconds())
			if index == 0 {
				for _, candidate := range result.Candidates {
					selected = append(selected, candidate.ID)
				}
			}
		}
		predictions[item.ID] = selected
	}

	byKindCases := map[string][]intermesheval.Case{}
	for _, item := range cases {
		kind := item.Kind
		if kind == "" {
			kind = "unspecified"
		}
		byKindCases[kind] = append(byKindCases[kind], item)
	}
	byKind := make(map[string]intermesheval.Metrics, len(byKindCases))
	for kind, items := range byKindCases {
		byKind[kind] = intermesheval.Score(items, predictions)
	}
	output := report{
		RegistryFingerprint: generation.Fingerprint,
		Metrics:             intermesheval.Score(cases, predictions),
		ByKind:              byKind,
		Latency: latencyReport{
			Samples: len(timings),
			P50:     intermesheval.Percentile(timings, 0.50),
			P95:     intermesheval.Percentile(timings, 0.95),
		},
		Predictions: predictions,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fatal(err)
	}
}

func evaluate(request route.Request, generation registry.Generation) (route.Result, error) {
	return graph.Resolve(route.Rank(request, generation), generation)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
