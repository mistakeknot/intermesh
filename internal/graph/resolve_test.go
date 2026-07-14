package graph

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
	"github.com/mistakeknot/intermesh/internal/skill"
)

func graphSkill(id string, requires, conflicts []string) skill.Skill {
	return skill.Skill{
		ID: id, Name: id, Description: "Description for " + id, SkillMD: "/skills/" + id + "/SKILL.md",
		Manifest: skill.Manifest{ID: id, Requires: requires, ConflictsWith: conflicts},
	}
}

func TestResolveOrdersTransitiveRequirementsBeforeSelectedSkill(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		graphSkill("fixture:a", []string{"fixture:b"}, nil),
		graphSkill("fixture:b", []string{"fixture:c"}, nil),
		graphSkill("fixture:c", nil, nil),
	}}
	input := route.Result{Candidates: []route.Candidate{{ID: "fixture:a", SelectedBy: "rank"}}}

	got, err := Resolve(input, generation)

	if err != nil {
		t.Fatal(err)
	}
	want := []string{"fixture:c", "fixture:b", "fixture:a"}
	if len(got.Candidates) != len(want) {
		t.Fatalf("unexpected candidate count: %#v", got.Candidates)
	}
	for index, id := range want {
		if got.Candidates[index].ID != id {
			t.Fatalf("unexpected order: %#v", got.Candidates)
		}
	}
	if got.Candidates[1].SelectedBy != "requirement" || got.Candidates[1].RequiredBy[0] != "fixture:a" {
		t.Fatalf("requirement provenance missing: %#v", got.Candidates[1])
	}
	payload, err := json.Marshal(got.Candidates[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"description":"Description for fixture:c"`) {
		t.Fatalf("required candidate description missing: %s", payload)
	}
	for _, candidate := range got.Candidates {
		payload, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(payload), `"description":"Description for `+candidate.ID+`"`) {
			t.Fatalf("candidate description was not hydrated: %s", payload)
		}
	}
}

func TestResolveDeduplicatesDiamondRequirements(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		graphSkill("fixture:a", []string{"fixture:b", "fixture:c"}, nil),
		graphSkill("fixture:b", []string{"fixture:d"}, nil),
		graphSkill("fixture:c", []string{"fixture:d"}, nil),
		graphSkill("fixture:d", nil, nil),
	}}
	input := route.Result{Candidates: []route.Candidate{{ID: "fixture:a"}}}

	got, err := Resolve(input, generation)

	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, candidate := range got.Candidates {
		if candidate.ID == "fixture:d" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("diamond dependency duplicated: %#v", got.Candidates)
	}
}

func TestResolveReportsExactCycle(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		graphSkill("fixture:a", []string{"fixture:b"}, nil),
		graphSkill("fixture:b", []string{"fixture:a"}, nil),
	}}

	_, err := Resolve(route.Result{Candidates: []route.Candidate{{ID: "fixture:a"}}}, generation)

	if err == nil || !strings.Contains(err.Error(), "fixture:a -> fixture:b -> fixture:a") {
		t.Fatalf("expected exact cycle, got %v", err)
	}
}

func TestResolveRejectsMissingRequirement(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{graphSkill("fixture:a", []string{"fixture:missing"}, nil)}}

	_, err := Resolve(route.Result{Candidates: []route.Candidate{{ID: "fixture:a"}}}, generation)

	if err == nil || !strings.Contains(err.Error(), "fixture:missing") {
		t.Fatalf("expected missing requirement error, got %v", err)
	}
}

func TestResolveSurfacesConflictWithoutDroppingCandidates(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		graphSkill("fixture:a", nil, []string{"fixture:b"}),
		graphSkill("fixture:b", nil, nil),
	}}
	input := route.Result{Candidates: []route.Candidate{{ID: "fixture:a"}, {ID: "fixture:b"}}}

	got, err := Resolve(input, generation)

	if err != nil {
		t.Fatal(err)
	}
	if len(got.Candidates) != 2 || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "conflicts") {
		t.Fatalf("conflict was not surfaced correctly: %#v", got)
	}
}
