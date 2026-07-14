package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "intermesh.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadParsesVersionOneManifest(t *testing.T) {
	path := writeManifest(t, `
version: 1
id: fixture:child
triggers:
  phrases: ["child fixture"]
  extensions: [".fixture"]
  environments: ["test"]
requires: ["fixture:parent"]
composes_with: ["fixture:peer"]
conflicts_with: ["fixture:conflict"]
supersedes: ["fixture:old"]
`)

	got, diagnostics := Load(path, "fallback:id")

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if got.ID != "fixture:child" || len(got.Requires) != 1 || got.Phrases[0] != "child fixture" {
		t.Fatalf("unexpected manifest: %#v", got)
	}
}

func TestLoadRejectsUnsupportedVersion(t *testing.T) {
	path := writeManifest(t, "version: 2\nid: fixture:test\n")

	_, diagnostics := Load(path, "")

	assertManifestDiagnostic(t, diagnostics, "manifest.unsupported_version")
}

func TestLoadRejectsPathLikeRelationshipID(t *testing.T) {
	path := writeManifest(t, "version: 1\nid: fixture:test\nrequires: [\"../escape\"]\n")

	_, diagnostics := Load(path, "")

	assertManifestDiagnostic(t, diagnostics, "manifest.invalid_relationship_id")
}

func assertManifestDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("diagnostic %q missing from %#v", code, diagnostics)
}
