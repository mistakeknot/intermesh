package skill

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxDescriptionLength = 1024

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func Parse(path, namespace string) (Skill, []Diagnostic) {
	canonical, err := filepath.Abs(path)
	if err != nil {
		return Skill{}, []Diagnostic{diagnostic(path, "path.absolute", err.Error())}
	}
	canonical = filepath.Clean(canonical)
	content, err := os.ReadFile(canonical)
	if err != nil {
		return Skill{}, []Diagnostic{diagnostic(canonical, "file.read", err.Error())}
	}

	normalized := strings.TrimPrefix(string(content), "\ufeff")
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Skill{}, []Diagnostic{diagnostic(canonical, "frontmatter.missing", "SKILL.md must start with ---")}
	}

	closing := -1
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			closing = index
			break
		}
	}
	if closing < 0 {
		return Skill{}, []Diagnostic{diagnostic(canonical, "frontmatter.unclosed", "frontmatter closing delimiter is missing")}
	}

	var metadata frontmatter
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:closing], "\n")), &metadata); err != nil {
		return Skill{}, []Diagnostic{diagnostic(canonical, "frontmatter.yaml", err.Error())}
	}
	metadata.Name = strings.TrimSpace(metadata.Name)
	metadata.Description = strings.TrimSpace(metadata.Description)
	body := strings.Join(lines[closing+1:], "\n")
	hash := sha256.Sum256([]byte(body))

	skill := Skill{
		Namespace:   strings.TrimSpace(namespace),
		Name:        metadata.Name,
		Description: metadata.Description,
		SkillMD:     canonical,
		Directory:   filepath.Dir(canonical),
		BodyHash:    fmt.Sprintf("sha256:%x", hash),
	}
	if skill.Namespace == "" {
		skill.ID = skill.Name
	} else {
		skill.ID = skill.Namespace + ":" + skill.Name
	}

	var diagnostics []Diagnostic
	if skill.Name == "" {
		diagnostics = append(diagnostics, diagnostic(canonical, "frontmatter.missing_name", "name is required"))
	} else if !validName.MatchString(skill.Name) {
		diagnostics = append(diagnostics, diagnostic(canonical, "frontmatter.invalid_name", "name must contain lowercase letters, digits, dot, underscore, or hyphen"))
	}
	if skill.Description == "" {
		diagnostics = append(diagnostics, diagnostic(canonical, "frontmatter.missing_description", "description is required"))
	} else if len([]rune(skill.Description)) > maxDescriptionLength {
		diagnostics = append(diagnostics, diagnostic(canonical, "frontmatter.description_too_long", "description exceeds the 1024-character host limit"))
	}
	return skill, diagnostics
}

func diagnostic(path, code, message string) Diagnostic {
	return Diagnostic{Path: path, Code: code, Message: message, Severity: "error"}
}
