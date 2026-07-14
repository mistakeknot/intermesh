package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mistakeknot/intermesh/internal/route"
)

func TestHelpListsCoreCommands(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := Run([]string{"--help"}, stdout, stderr)

	if exit != ExitOK || !strings.Contains(stdout.String(), "route") || !strings.Contains(stdout.String(), "doctor") {
		t.Fatalf("unexpected help exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
}

func TestIndexRouteResolveJSONHappyPath(t *testing.T) {
	root := t.TempDir()
	copyCLIFixture(t, filepath.Join("..", "..", "testdata", "skills", "minimal"), filepath.Join(root, "minimal"))
	copyCLIFixture(t, filepath.Join("..", "..", "testdata", "skills", "depends"), filepath.Join(root, "depends"))
	database := filepath.Join(t.TempDir(), "registry.db")

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	exit := Run([]string{"index", "--root", "fixture=" + root, "--db", database, "--json"}, stdout, stderr)
	if exit != ExitOK {
		t.Fatalf("index failed exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}

	stdout.Reset()
	stderr.Reset()
	exit = Run([]string{"route", "--query", "dependent fixture", "--host", "codex", "--db", database, "--limit", "1", "--no-receipt", "--json"}, stdout, stderr)
	if exit != ExitOK {
		t.Fatalf("route failed exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
	var result route.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 || result.Candidates[0].ID != "fixture:minimal" || result.Candidates[1].ID != "fixture:depends" {
		t.Fatalf("route did not expand requirement in load order: %#v", result.Candidates)
	}
}

func TestDoctorReturnsUnhealthyForEmptyRegistry(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := Run([]string{"doctor", "--db", filepath.Join(t.TempDir(), "registry.db"), "--json"}, stdout, stderr)

	if exit != ExitRegistryUnhealthy || !strings.Contains(stdout.String(), `"healthy":false`) {
		t.Fatalf("unexpected doctor exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
}

func TestInvalidInvocationReturnsUsageExit(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := Run([]string{"route", "--db", filepath.Join(t.TempDir(), "registry.db")}, stdout, stderr)

	if exit != ExitInvalidInput || !strings.Contains(stderr.String(), "query") {
		t.Fatalf("unexpected invalid invocation exit=%d stderr=%q", exit, stderr)
	}
}

func copyCLIFixture(t *testing.T, source, destination string) {
	t.Helper()
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destination, entry.Name()), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
