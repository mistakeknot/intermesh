package registry

import (
	"context"
	"os"

	"github.com/mistakeknot/intermesh/internal/manifest"
	"github.com/mistakeknot/intermesh/internal/skill"
)

func Index(ctx context.Context, store *Store, roots []skill.Root) (Generation, error) {
	discovered, diagnostics, err := skill.Discover(ctx, roots)
	if err != nil {
		return Generation{}, err
	}
	withManifests := make([]skill.Skill, 0, len(discovered))
	for _, item := range discovered {
		path := manifest.PathForSkill(item)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				withManifests = append(withManifests, item)
				continue
			}
			return Generation{}, err
		}
		parsed, manifestDiagnostics := manifest.Load(path, item.ID)
		diagnostics = append(diagnostics, manifestDiagnostics...)
		invalid := false
		for _, diagnostic := range manifestDiagnostics {
			if diagnostic.Severity == "error" {
				invalid = true
				break
			}
		}
		if invalid {
			continue
		}
		item.Manifest = parsed
		item.ID = parsed.ID
		withManifests = append(withManifests, item)
	}
	return store.Replace(ctx, withManifests, roots, diagnostics)
}
