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

func TestProfilePlanWritesPlanWithoutMutatingCatalog(t *testing.T) {
	catalog := t.TempDir()
	source := filepath.Join(t.TempDir(), "source")
	router := filepath.Join(t.TempDir(), "router")
	if err := os.MkdirAll(source, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(router, 0o700); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(catalog, "source")
	if err := os.Symlink(source, entry); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := Run([]string{"profile", "plan", "--host", "codex", "--catalog", catalog, "--router", router, "--out", planPath, "--json"}, stdout, stderr)

	if exit != ExitOK || !strings.Contains(stdout.String(), `"mode":"managed"`) {
		t.Fatalf("unexpected plan exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
	if _, err := os.Lstat(entry); err != nil {
		t.Fatalf("planning mutated catalog: %v", err)
	}
	if content, err := os.ReadFile(planPath); err != nil || !strings.Contains(string(content), `"host":"codex"`) {
		t.Fatalf("plan file missing or invalid content=%q err=%v", content, err)
	}
}

func TestProfileApplyAndRestoreRequireExplicitConfirmation(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := Run([]string{"profile", "apply", "--plan", "unused.json"}, stdout, stderr)
	if exit != ExitUnsafeProfile || !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("apply confirmation not enforced exit=%d stderr=%q", exit, stderr)
	}
	stdout.Reset()
	stderr.Reset()
	exit = Run([]string{"profile", "restore", "--snapshot", "unused.json"}, stdout, stderr)
	if exit != ExitUnsafeProfile || !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("restore confirmation not enforced exit=%d stderr=%q", exit, stderr)
	}
}

func TestProfileCLIApplyAndRestoreRoundTrip(t *testing.T) {
	catalog := t.TempDir()
	source := filepath.Join(t.TempDir(), "source")
	router := filepath.Join(t.TempDir(), "router")
	if err := os.MkdirAll(source, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(router, 0o700); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(catalog, "source")
	if err := os.Symlink(source, entry); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	snapshotRoot := t.TempDir()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if exit := Run([]string{"profile", "plan", "--host", "codex", "--catalog", catalog, "--router", router, "--out", planPath}, stdout, stderr); exit != ExitOK {
		t.Fatalf("plan failed exit=%d stderr=%q", exit, stderr)
	}
	stdout.Reset()
	stderr.Reset()
	if exit := Run([]string{"profile", "apply", "--plan", planPath, "--snapshot-root", snapshotRoot, "--yes", "--json"}, stdout, stderr); exit != ExitOK {
		t.Fatalf("apply failed exit=%d stderr=%q", exit, stderr)
	}
	var applied struct {
		Snapshot string `json:"snapshot"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil || applied.Snapshot == "" {
		t.Fatalf("missing snapshot output=%q err=%v", stdout, err)
	}
	stdout.Reset()
	stderr.Reset()
	if exit := Run([]string{"profile", "restore", "--snapshot", applied.Snapshot, "--yes"}, stdout, stderr); exit != ExitOK {
		t.Fatalf("restore failed exit=%d stderr=%q", exit, stderr)
	}
	if target, err := os.Readlink(entry); err != nil || target != source {
		t.Fatalf("entry not restored target=%q err=%v", target, err)
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
