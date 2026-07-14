package registry

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mistakeknot/intermesh/internal/skill"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func fixtureSkill(id, path string) skill.Skill {
	return skill.Skill{
		ID:          id,
		Namespace:   "fixture",
		Name:        id,
		Description: "Use for " + id,
		SkillMD:     path,
		Directory:   filepath.Dir(path),
		BodyHash:    "sha256:" + id,
	}
}

func TestReplaceProducesDeterministicFingerprint(t *testing.T) {
	ctx := context.Background()
	firstStore := openTestStore(t)
	secondStore := openTestStore(t)
	a := fixtureSkill("a", "/skills/a/SKILL.md")
	b := fixtureSkill("b", "/skills/b/SKILL.md")

	first, err := firstStore.Replace(ctx, []skill.Skill{a, b}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := secondStore.Replace(ctx, []skill.Skill{b, a}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Fingerprint != second.Fingerprint {
		t.Fatalf("fingerprint depends on input order: %q != %q", first.Fingerprint, second.Fingerprint)
	}
}

func TestReplaceDuplicateIDPreservesPreviousGeneration(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	previous, err := store.Replace(ctx, []skill.Skill{fixtureSkill("stable", "/skills/stable/SKILL.md")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Replace(ctx, []skill.Skill{
		fixtureSkill("duplicate", "/skills/one/SKILL.md"),
		fixtureSkill("duplicate", "/skills/two/SKILL.md"),
	}, nil, nil)
	if err == nil {
		t.Fatal("expected duplicate ID error")
	}

	current, err := store.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if current.Fingerprint != previous.Fingerprint || len(current.Skills) != 1 || current.Skills[0].ID != "stable" {
		t.Fatalf("previous generation was not preserved: %#v", current)
	}
}

func TestReplaceRemovesSkillsMissingFromNextGeneration(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	if _, err := store.Replace(ctx, []skill.Skill{
		fixtureSkill("keep", "/skills/keep/SKILL.md"),
		fixtureSkill("remove", "/skills/remove/SKILL.md"),
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Replace(ctx, []skill.Skill{fixtureSkill("keep", "/skills/keep/SKILL.md")}, nil, nil); err != nil {
		t.Fatal(err)
	}

	current, err := store.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Skills) != 1 || current.Skills[0].ID != "keep" {
		t.Fatalf("unexpected current generation: %#v", current.Skills)
	}
}
