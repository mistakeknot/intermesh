package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildManagedPlanForSymlinkCatalog(t *testing.T) {
	catalog := t.TempDir()
	source := filepath.Join(t.TempDir(), "source-skill")
	router := filepath.Join(t.TempDir(), "router")
	mustMkdir(t, source)
	mustMkdir(t, router)
	if err := os.Symlink(source, filepath.Join(catalog, "source-skill")); err != nil {
		t.Fatal(err)
	}

	plan, err := Build("codex", catalog, router, nil)

	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != ModeManaged || len(plan.Entries) != 1 || plan.Entries[0].Target != source {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if _, err := os.Lstat(filepath.Join(catalog, "source-skill")); err != nil {
		t.Fatalf("planning mutated catalog: %v", err)
	}
}

func TestBuildMarksRegularDirectoryCatalogRoutingOnly(t *testing.T) {
	catalog := t.TempDir()
	router := filepath.Join(t.TempDir(), "router")
	mustMkdir(t, filepath.Join(catalog, "user-authored"))
	mustMkdir(t, router)

	plan, err := Build("hermes", catalog, router, nil)

	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != ModeRoutingOnly || len(plan.Blockers) == 0 {
		t.Fatalf("regular directory should block managed mode: %#v", plan)
	}
}

func TestApplyAndRestoreRoundTripSymlinks(t *testing.T) {
	catalog := t.TempDir()
	source := filepath.Join(t.TempDir(), "source-skill")
	router := filepath.Join(t.TempDir(), "router")
	mustMkdir(t, source)
	mustMkdir(t, router)
	original := filepath.Join(catalog, "source-skill")
	if err := os.Symlink(source, original); err != nil {
		t.Fatal(err)
	}
	plan, err := Build("codex", catalog, router, nil)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := Apply(plan, t.TempDir())

	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(original); !os.IsNotExist(err) {
		t.Fatalf("original catalog link still exists: %v", err)
	}
	routerLink := filepath.Join(catalog, RouterEntryName)
	if target, err := os.Readlink(routerLink); err != nil || target != plan.Router {
		t.Fatalf("router link missing target=%q err=%v", target, err)
	}
	if _, err := os.Stat(snapshot.Path); err != nil {
		t.Fatalf("snapshot missing: %v", err)
	}
	if !strings.HasPrefix(snapshot.Fingerprint, "sha256:") {
		t.Fatalf("snapshot missing integrity fingerprint: %#v", snapshot)
	}

	if err := Restore(snapshot.Path); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(original); err != nil || target != source {
		t.Fatalf("original link not restored target=%q err=%v", target, err)
	}
	if _, err := os.Lstat(routerLink); !os.IsNotExist(err) {
		t.Fatalf("router link was not removed: %v", err)
	}
}

func TestApplyIsIdempotentForAlreadyActivePlan(t *testing.T) {
	catalog, _, _, plan := managedFixture(t)

	first, err := Apply(plan, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	second, err := Apply(plan, t.TempDir())
	if err != nil {
		t.Fatalf("repeat activation should succeed: %v", err)
	}
	if first.Fingerprint != second.Fingerprint {
		t.Fatalf("repeat activation changed snapshot identity: first=%q second=%q", first.Fingerprint, second.Fingerprint)
	}
	if target, err := os.Readlink(filepath.Join(catalog, RouterEntryName)); err != nil || target != plan.Router {
		t.Fatalf("router profile drifted target=%q err=%v", target, err)
	}
}

func TestRestoreRecoversSnapshotWrittenBeforeAnyMutation(t *testing.T) {
	_, source, original, plan := managedFixture(t)
	snapshot, err := snapshotForPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	path, err := writeSnapshot(t.TempDir(), snapshot)
	if err != nil {
		t.Fatal(err)
	}

	if err := Restore(path); err != nil {
		t.Fatalf("snapshot-only recovery should be a no-op: %v", err)
	}
	if target, err := os.Readlink(original); err != nil || target != source {
		t.Fatalf("original changed target=%q err=%v", target, err)
	}
}

func TestApplyRefusesEntryOutsideCatalog(t *testing.T) {
	_, _, _, plan := managedFixture(t)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.Symlink(plan.Entries[0].Target, outside); err != nil {
		t.Fatal(err)
	}
	plan.Entries[0].Path = outside

	if _, err := Apply(plan, t.TempDir()); err == nil || !strings.Contains(err.Error(), "escapes catalog") {
		t.Fatalf("expected traversal refusal, got %v", err)
	}
}

func TestApplyRefusesCatalogDrift(t *testing.T) {
	catalog := t.TempDir()
	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(t.TempDir(), "second")
	router := filepath.Join(t.TempDir(), "router")
	mustMkdir(t, first)
	mustMkdir(t, second)
	mustMkdir(t, router)
	entry := filepath.Join(catalog, "skill")
	if err := os.Symlink(first, entry); err != nil {
		t.Fatal(err)
	}
	plan, err := Build("codex", catalog, router, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(entry); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(second, entry); err != nil {
		t.Fatal(err)
	}

	if _, err := Apply(plan, t.TempDir()); err == nil {
		t.Fatal("expected drift refusal")
	}
	if target, err := os.Readlink(entry); err != nil || target != second {
		t.Fatalf("drifted entry was mutated target=%q err=%v", target, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func managedFixture(t *testing.T) (catalog, source, original string, plan Plan) {
	t.Helper()
	catalog = t.TempDir()
	source = filepath.Join(t.TempDir(), "source")
	router := filepath.Join(t.TempDir(), "router")
	mustMkdir(t, source)
	mustMkdir(t, router)
	original = filepath.Join(catalog, "source")
	if err := os.Symlink(source, original); err != nil {
		t.Fatal(err)
	}
	var err error
	plan, err = Build("codex", catalog, router, nil)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, source, original, plan
}
