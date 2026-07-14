package registry

import (
	"context"

	"github.com/mistakeknot/intermesh/internal/skill"
)

func Index(ctx context.Context, store *Store, roots []skill.Root) (Generation, error) {
	discovered, diagnostics, err := skill.Discover(ctx, roots)
	if err != nil {
		return Generation{}, err
	}
	return store.Replace(ctx, discovered, roots, diagnostics)
}
