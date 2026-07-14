package skill

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func Discover(ctx context.Context, roots []Root) ([]Skill, []Diagnostic, error) {
	normalizedRoots, diagnostics := normalizeRoots(roots)
	seenPaths := make(map[string]struct{})
	var skills []Skill

	for _, root := range normalizedRoots {
		err := filepath.WalkDir(root.Path, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				diagnostics = append(diagnostics, diagnostic(path, "path.walk", walkErr.Error()))
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if entry.IsDir() || entry.Name() != "SKILL.md" {
				return nil
			}

			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				diagnostics = append(diagnostics, diagnostic(path, "path.resolve", err.Error()))
				return nil
			}
			resolved, err = filepath.Abs(resolved)
			if err != nil {
				diagnostics = append(diagnostics, diagnostic(path, "path.absolute", err.Error()))
				return nil
			}
			resolved = filepath.Clean(resolved)
			if !withinRoot(root.Path, resolved) {
				diagnostics = append(diagnostics, diagnostic(path, "path.outside_root", fmt.Sprintf("resolved path %s escapes root %s", resolved, root.Path)))
				return nil
			}
			if _, exists := seenPaths[resolved]; exists {
				return nil
			}
			seenPaths[resolved] = struct{}{}

			skill, parsedDiagnostics := Parse(resolved, root.Namespace)
			diagnostics = append(diagnostics, parsedDiagnostics...)
			if !hasErrors(parsedDiagnostics) {
				skills = append(skills, skill)
			}
			return nil
		})
		if err != nil {
			return nil, diagnostics, err
		}
	}

	sort.Slice(skills, func(i, j int) bool { return skills[i].SkillMD < skills[j].SkillMD })
	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].Path == diagnostics[j].Path {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return diagnostics[i].Path < diagnostics[j].Path
	})
	return skills, diagnostics, nil
}

func normalizeRoots(roots []Root) ([]Root, []Diagnostic) {
	seen := make(map[string]struct{})
	var normalized []Root
	var diagnostics []Diagnostic
	for _, root := range roots {
		absolute, err := filepath.Abs(root.Path)
		if err != nil {
			diagnostics = append(diagnostics, diagnostic(root.Path, "root.absolute", err.Error()))
			continue
		}
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			diagnostics = append(diagnostics, diagnostic(root.Path, "root.resolve", err.Error()))
			continue
		}
		resolved = filepath.Clean(resolved)
		key := resolved + "\x00" + strings.TrimSpace(root.Namespace)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, Root{Path: resolved, Namespace: strings.TrimSpace(root.Namespace)})
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Path == normalized[j].Path {
			return normalized[i].Namespace < normalized[j].Namespace
		}
		return normalized[i].Path < normalized[j].Path
	})
	return normalized, diagnostics
}

func withinRoot(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
