package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func Build(host, catalog, router string, alwaysOn []string) (Plan, error) {
	catalogPath, err := canonicalDirectory(catalog)
	if err != nil {
		return Plan{}, fmt.Errorf("catalog: %w", err)
	}
	routerPath, err := canonicalDirectory(router)
	if err != nil {
		return Plan{}, fmt.Errorf("router: %w", err)
	}
	plan := Plan{Version: 1, Host: host, Catalog: catalogPath, Router: routerPath, Mode: ModeManaged, AlwaysOn: append([]string(nil), alwaysOn...), Entries: []Entry{}, Blockers: []string{}}
	always := make(map[string]struct{}, len(alwaysOn))
	for _, name := range alwaysOn {
		always[name] = struct{}{}
	}
	entries, err := os.ReadDir(catalogPath)
	if err != nil {
		return Plan{}, err
	}
	for _, directoryEntry := range entries {
		name := directoryEntry.Name()
		if name == RouterEntryName {
			continue
		}
		if _, keep := always[name]; keep {
			continue
		}
		path := filepath.Join(catalogPath, name)
		info, err := os.Lstat(path)
		if err != nil {
			return Plan{}, err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			plan.Mode = ModeRoutingOnly
			plan.Blockers = append(plan.Blockers, fmt.Sprintf("%s is not a symlink; V0 never moves or deletes regular catalog entries", path))
			continue
		}
		target, err := os.Readlink(path)
		if err != nil {
			return Plan{}, err
		}
		plan.Entries = append(plan.Entries, Entry{Name: name, Path: path, Target: target})
	}
	sort.Slice(plan.Entries, func(i, j int) bool { return plan.Entries[i].Name < plan.Entries[j].Name })
	sort.Strings(plan.AlwaysOn)
	sort.Strings(plan.Blockers)
	return plan, nil
}

func canonicalDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", resolved)
	}
	return filepath.Clean(resolved), nil
}
