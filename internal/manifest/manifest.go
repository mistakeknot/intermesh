package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mistakeknot/intermesh/internal/skill"
	"gopkg.in/yaml.v3"
)

type Diagnostic = skill.Diagnostic

var validID = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*(?::[a-z0-9][a-z0-9._-]*)?$`)

type rawManifest struct {
	Version  int    `yaml:"version"`
	ID       string `yaml:"id"`
	Triggers struct {
		Phrases      []string `yaml:"phrases"`
		Extensions   []string `yaml:"extensions"`
		Environments []string `yaml:"environments"`
	} `yaml:"triggers"`
	Requires      []string `yaml:"requires"`
	ComposesWith  []string `yaml:"composes_with"`
	ConflictsWith []string `yaml:"conflicts_with"`
	Supersedes    []string `yaml:"supersedes"`
}

func Load(path, fallbackID string) (skill.Manifest, []Diagnostic) {
	content, err := os.ReadFile(path)
	if err != nil {
		return skill.Manifest{}, []Diagnostic{manifestDiagnostic(path, "manifest.read", err.Error())}
	}
	var raw rawManifest
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return skill.Manifest{}, []Diagnostic{manifestDiagnostic(path, "manifest.yaml", err.Error())}
	}
	if raw.Version != 1 {
		return skill.Manifest{}, []Diagnostic{manifestDiagnostic(path, "manifest.unsupported_version", fmt.Sprintf("version %d is not supported", raw.Version))}
	}
	identifier := strings.TrimSpace(raw.ID)
	if identifier == "" {
		identifier = fallbackID
	}
	var diagnostics []Diagnostic
	if !validID.MatchString(identifier) {
		diagnostics = append(diagnostics, manifestDiagnostic(path, "manifest.invalid_id", fmt.Sprintf("invalid skill ID %q", identifier)))
	}
	relations := map[string][]string{
		"requires": raw.Requires, "composes_with": raw.ComposesWith,
		"conflicts_with": raw.ConflictsWith, "supersedes": raw.Supersedes,
	}
	for relation, identifiers := range relations {
		for _, target := range identifiers {
			if !validID.MatchString(strings.TrimSpace(target)) {
				diagnostics = append(diagnostics, manifestDiagnostic(path, "manifest.invalid_relationship_id", fmt.Sprintf("%s contains invalid skill ID %q", relation, target)))
			}
		}
	}
	if hasErrors(diagnostics) {
		return skill.Manifest{}, diagnostics
	}
	manifest := skill.Manifest{
		Version: raw.Version, ID: identifier,
		Phrases: clean(raw.Triggers.Phrases), Extensions: clean(raw.Triggers.Extensions), Environments: clean(raw.Triggers.Environments),
		Requires: clean(raw.Requires), ComposesWith: clean(raw.ComposesWith),
		ConflictsWith: clean(raw.ConflictsWith), Supersedes: clean(raw.Supersedes),
	}
	return manifest, diagnostics
}

func PathForSkill(item skill.Skill) string {
	return filepath.Join(item.Directory, "intermesh.yaml")
}

func clean(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func manifestDiagnostic(path, code, message string) Diagnostic {
	return Diagnostic{Path: path, Code: code, Message: message, Severity: "error"}
}

func hasErrors(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return true
		}
	}
	return false
}
