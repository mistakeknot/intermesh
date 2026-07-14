package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mistakeknot/intermesh/internal/skill"
)

func TestIndexDiscoversValidSkillsAndPersistsDiagnostics(t *testing.T) {
	root := t.TempDir()
	writeIndexSkill(t, filepath.Join(root, "valid", "SKILL.md"), "valid", true)
	writeIndexSkill(t, filepath.Join(root, "invalid", "SKILL.md"), "invalid", false)
	store := openTestStore(t)

	generation, err := Index(context.Background(), store, []skill.Root{{Path: root, Namespace: "fixture"}})

	if err != nil {
		t.Fatal(err)
	}
	if generation.SkillCount != 1 || generation.DiagnosticCount != 1 {
		t.Fatalf("unexpected generation counts: %#v", generation)
	}
	current, err := store.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Diagnostics) != 1 || current.Diagnostics[0].Code != "frontmatter.missing_description" {
		t.Fatalf("diagnostics were not persisted: %#v", current.Diagnostics)
	}
}

func writeIndexSkill(t *testing.T, path, name string, withDescription bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\n"
	if withDescription {
		content += "description: Use for " + name + ".\n"
	}
	content += "---\nBody\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
