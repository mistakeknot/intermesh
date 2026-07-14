package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/mistakeknot/intermesh/internal/registry"
)

type doctorResult struct {
	Version         int    `json:"version"`
	Healthy         bool   `json:"healthy"`
	Database        string `json:"database"`
	Fingerprint     string `json:"fingerprint,omitempty"`
	SkillCount      int    `json:"skill_count"`
	DiagnosticCount int    `json:"diagnostic_count"`
	Message         string `json:"message,omitempty"`
}

func runDoctor(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	database := flags.String("db", registry.DefaultPath(), "registry database")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	result := doctorResult{Version: 1, Database: *database}
	store, err := registry.Open(*database)
	if err == nil {
		defer store.Close()
		generation, currentErr := store.Current(context.Background())
		if currentErr == nil && generation.Fingerprint != "" {
			result.Healthy = true
			result.Fingerprint = generation.Fingerprint
			result.SkillCount = generation.SkillCount
			result.DiagnosticCount = generation.DiagnosticCount
		} else if currentErr != nil {
			err = currentErr
		}
	}
	if !result.Healthy {
		if err != nil {
			result.Message = err.Error()
		} else {
			result.Message = "registry is empty; run intermesh index"
		}
	}
	if *jsonOutput {
		_ = writeJSON(stdout, result)
	} else if result.Healthy {
		fmt.Fprintf(stdout, "healthy: %d skills %s\n", result.SkillCount, result.Fingerprint)
	} else {
		fmt.Fprintf(stderr, "unhealthy: %s\n", result.Message)
	}
	if result.Healthy {
		return ExitOK
	}
	return ExitRegistryUnhealthy
}
