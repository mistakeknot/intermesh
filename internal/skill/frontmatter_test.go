package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkillFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseReadsFrontmatterAndBodyHash(t *testing.T) {
	path := writeSkillFile(t, "---\nname: sample\ndescription: Use for sample work.\n---\n\n# Sample\n\nDo the work.\n")

	got, diagnostics := Parse(path, "fixture")

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if got.ID != "fixture:sample" || got.Name != "sample" {
		t.Fatalf("unexpected identity: %#v", got)
	}
	if got.Description != "Use for sample work." {
		t.Fatalf("unexpected description: %q", got.Description)
	}
	if !strings.HasPrefix(got.BodyHash, "sha256:") {
		t.Fatalf("unexpected body hash: %q", got.BodyHash)
	}
}

func TestParseAcceptsBOMCRLFAndFoldedDescription(t *testing.T) {
	path := writeSkillFile(t, "\ufeff---\r\nname: folded\r\ndescription: >-\r\n  Use when a description\r\n  spans lines.\r\n---\r\nBody\r\n")

	got, diagnostics := Parse(path, "")

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if got.ID != "folded" {
		t.Fatalf("unexpected ID: %q", got.ID)
	}
	if got.Description != "Use when a description spans lines." {
		t.Fatalf("unexpected folded description: %q", got.Description)
	}
}

func TestParseReportsMissingClosingDelimiter(t *testing.T) {
	path := writeSkillFile(t, "---\nname: broken\ndescription: broken\n")

	_, diagnostics := Parse(path, "")

	assertDiagnosticCode(t, diagnostics, "frontmatter.unclosed")
}

func TestParseReportsDescriptionOverHostLimit(t *testing.T) {
	path := writeSkillFile(t, "---\nname: verbose\ndescription: \""+strings.Repeat("x", 1025)+"\"\n---\nBody\n")

	_, diagnostics := Parse(path, "")

	assertDiagnosticCode(t, diagnostics, "frontmatter.description_too_long")
}

func assertDiagnosticCode(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("diagnostic %q not found in %#v", code, diagnostics)
}
