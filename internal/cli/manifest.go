package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/mistakeknot/intermesh/internal/manifest"
)

func runManifest(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 || arguments[0] != "validate" {
		fmt.Fprintln(stderr, "usage: intermesh manifest validate path [--json]")
		return ExitInvalidInput
	}
	flags := flag.NewFlagSet("manifest validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "emit JSON")
	if err := flags.Parse(arguments[1:]); err != nil {
		return ExitInvalidInput
	}
	if len(flags.Args()) != 1 {
		fmt.Fprintln(stderr, "manifest path is required")
		return ExitInvalidInput
	}
	parsed, diagnostics := manifest.Load(flags.Args()[0], "")
	if *jsonOutput {
		_ = writeJSON(stdout, map[string]any{"valid": len(diagnostics) == 0, "manifest": parsed, "diagnostics": diagnostics})
	} else if len(diagnostics) == 0 {
		fmt.Fprintf(stdout, "valid manifest %s\n", parsed.ID)
	}
	if len(diagnostics) > 0 {
		for _, diagnostic := range diagnostics {
			fmt.Fprintf(stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
		}
		return ExitInvalidInput
	}
	return ExitOK
}
