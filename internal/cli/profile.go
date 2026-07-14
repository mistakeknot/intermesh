package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mistakeknot/intermesh/internal/profile"
)

func runProfile(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		fmt.Fprintln(stderr, "usage: intermesh profile <plan|apply|restore> [flags]")
		return ExitInvalidInput
	}
	switch arguments[0] {
	case "plan":
		return runProfilePlan(arguments[1:], stdout, stderr)
	case "apply":
		return runProfileApply(arguments[1:], stdout, stderr)
	case "restore":
		return runProfileRestore(arguments[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown profile command %q\n", arguments[0])
		return ExitInvalidInput
	}
}

func runProfilePlan(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("profile plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	host := flags.String("host", "", "agent host")
	catalog := flags.String("catalog", "", "host skill catalog")
	router := flags.String("router", "", "router skill directory")
	output := flags.String("out", "", "write plan JSON to file")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	var alwaysOn stringList
	flags.Var(&alwaysOn, "always-on", "catalog entry to preserve; repeatable")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if strings.TrimSpace(*host) == "" || strings.TrimSpace(*catalog) == "" || strings.TrimSpace(*router) == "" {
		fmt.Fprintln(stderr, "--host, --catalog, and --router are required")
		return ExitInvalidInput
	}
	plan, err := profile.Build(*host, *catalog, *router, alwaysOn)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitInvalidInput
	}
	if *output != "" {
		if err := writeJSONFile(*output, plan); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitInvalidInput
		}
	}
	if *jsonOutput {
		if err := writeJSON(stdout, plan); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitInvalidInput
		}
	} else {
		fmt.Fprintf(stdout, "profile plan: %s (%d removable entries)\n", plan.Mode, len(plan.Entries))
		for _, blocker := range plan.Blockers {
			fmt.Fprintf(stdout, "blocker: %s\n", blocker)
		}
		if *output != "" {
			fmt.Fprintf(stdout, "wrote %s\n", *output)
		}
	}
	return ExitOK
}

func runProfileApply(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("profile apply", flag.ContinueOnError)
	flags.SetOutput(stderr)
	planPath := flags.String("plan", "", "profile plan JSON")
	snapshotRoot := flags.String("snapshot-root", defaultSnapshotRoot(), "snapshot directory")
	yes := flags.Bool("yes", false, "confirm catalog mutation")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if !*yes {
		fmt.Fprintln(stderr, "refusing to mutate a skill catalog without --yes")
		return ExitUnsafeProfile
	}
	if *planPath == "" {
		fmt.Fprintln(stderr, "--plan is required")
		return ExitInvalidInput
	}
	var plan profile.Plan
	if err := readJSONFile(*planPath, &plan); err != nil {
		fmt.Fprintln(stderr, err)
		return ExitInvalidInput
	}
	snapshot, err := profile.Apply(plan, *snapshotRoot)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUnsafeProfile
	}
	result := map[string]string{"snapshot": snapshot.Path, "mode": plan.Mode}
	if *jsonOutput {
		_ = writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "activated router profile; snapshot %s\n", snapshot.Path)
	}
	return ExitOK
}

func runProfileRestore(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("profile restore", flag.ContinueOnError)
	flags.SetOutput(stderr)
	snapshot := flags.String("snapshot", "", "profile snapshot JSON")
	yes := flags.Bool("yes", false, "confirm catalog mutation")
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments); err != nil {
		return ExitInvalidInput
	}
	if !*yes {
		fmt.Fprintln(stderr, "refusing to mutate a skill catalog without --yes")
		return ExitUnsafeProfile
	}
	if *snapshot == "" {
		fmt.Fprintln(stderr, "--snapshot is required")
		return ExitInvalidInput
	}
	if err := profile.Restore(*snapshot); err != nil {
		fmt.Fprintln(stderr, err)
		return ExitUnsafeProfile
	}
	if *jsonOutput {
		_ = writeJSON(stdout, map[string]string{"restored": *snapshot})
	} else {
		fmt.Fprintf(stdout, "restored profile snapshot %s\n", *snapshot)
	}
	return ExitOK
}

func writeJSONFile(path string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o600)
}

func readJSONFile(path string, value any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, value); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func defaultSnapshotRoot() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "state", "intermesh", "profiles")
	}
	return filepath.Join(os.TempDir(), "intermesh-profiles")
}
