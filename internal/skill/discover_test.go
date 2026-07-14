package skill

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscoverFindsValidSkillsAndOmitsInvalidOnes(t *testing.T) {
	root := t.TempDir()
	copyFixtureTree(t, filepath.Join("..", "..", "testdata", "skills"), root)

	skills, diagnostics, err := Discover(context.Background(), []Root{{Path: root, Namespace: "fixture"}})

	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 3 {
		t.Fatalf("expected 3 valid skills, got %d: %#v", len(skills), skills)
	}
	assertDiagnosticCode(t, diagnostics, "frontmatter.missing_description")
	for i := 1; i < len(skills); i++ {
		if skills[i-1].SkillMD > skills[i].SkillMD {
			t.Fatalf("skills are not path-sorted: %#v", skills)
		}
	}
}

func TestDiscoverDeduplicatesCanonicalRoots(t *testing.T) {
	root := t.TempDir()
	writeSkillAt(t, filepath.Join(root, "one", "SKILL.md"), "one")

	skills, _, err := Discover(context.Background(), []Root{{Path: root}, {Path: filepath.Join(root, ".")}})

	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected one deduplicated skill, got %d", len(skills))
	}
}

func TestDiscoverRejectsSymlinkEscapingRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	writeSkillAt(t, filepath.Join(outside, "SKILL.md"), "outside")
	if err := os.Symlink(filepath.Join(outside, "SKILL.md"), filepath.Join(root, "SKILL.md")); err != nil {
		t.Fatal(err)
	}

	skills, diagnostics, err := Discover(context.Background(), []Root{{Path: root}})

	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Fatalf("expected escaped symlink to be omitted, got %#v", skills)
	}
	assertDiagnosticCode(t, diagnostics, "path.outside_root")
}

func writeSkillAt(t *testing.T, path, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: Use for " + name + ".\n---\nBody\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func copyFixtureTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
}
