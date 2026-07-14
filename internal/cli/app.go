package cli

import (
	"fmt"
	"io"
)

const (
	ExitOK                = 0
	ExitInvalidInput      = 2
	ExitRegistryUnhealthy = 3
	ExitGraphUnresolved   = 4
	ExitUnsafeProfile     = 5
)

func Run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 || arguments[0] == "--help" || arguments[0] == "-h" || arguments[0] == "help" {
		writeHelp(stdout)
		return ExitOK
	}
	command := arguments[0]
	arguments = arguments[1:]
	switch command {
	case "index":
		return runIndex(arguments, stdout, stderr)
	case "route":
		return runRoute(arguments, stdout, stderr)
	case "search":
		return runSearch(arguments, stdout, stderr)
	case "resolve":
		return runResolve(arguments, stdout, stderr)
	case "graph":
		return runGraph(arguments, stdout, stderr)
	case "manifest":
		return runManifest(arguments, stdout, stderr)
	case "doctor":
		return runDoctor(arguments, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", command)
		writeHelp(stderr)
		return ExitInvalidInput
	}
}

func writeHelp(writer io.Writer) {
	fmt.Fprintln(writer, `intermesh - retrieval-gated Agent Skills discovery

Usage:
  intermesh index --root [namespace=]path [--db path] [--json]
  intermesh route --query text [--host host] [--limit 5] [--json]
  intermesh search [flags] query
  intermesh resolve [flags] id...
  intermesh graph [flags] id
  intermesh manifest validate path [--json]
  intermesh doctor [--db path] [--json]`)
}
