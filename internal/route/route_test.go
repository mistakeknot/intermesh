package route

import (
	"math"
	"testing"

	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/skill"
)

func routeFixture(id, name, description string, manifest skill.Manifest) skill.Skill {
	return skill.Skill{
		ID: id, Name: name, Description: description,
		SkillMD: "/skills/" + name + "/SKILL.md", Manifest: manifest,
	}
}

func TestRankExactIDWins(t *testing.T) {
	generation := registry.Generation{Fingerprint: "sha256:test", Skills: []skill.Skill{
		routeFixture("intertest:verify", "verify", "Verify completed engineering work.", skill.Manifest{}),
		routeFixture("intertest:test", "test", "Test engineering work.", skill.Manifest{}),
	}}

	result := Rank(Request{Query: "use intertest:verify", Host: "codex", Limit: 5}, generation)

	if len(result.Candidates) == 0 || result.Candidates[0].ID != "intertest:verify" {
		t.Fatalf("exact ID did not win: %#v", result.Candidates)
	}
	assertReason(t, result.Candidates[0], "exact_id")
	if result.Candidates[0].Components["exact_id"] != 100 {
		t.Fatalf("exact score component missing: %#v", result.Candidates[0].Components)
	}
	var componentTotal float64
	for _, value := range result.Candidates[0].Components {
		componentTotal += value
	}
	if math.Abs(componentTotal-result.Candidates[0].Score) > 1e-9 {
		t.Fatalf("components do not sum to score: components=%#v score=%f", result.Candidates[0].Components, result.Candidates[0].Score)
	}
}

func TestRankUsesPhraseExtensionEnvironmentAndLexicalSignals(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("docs:pdf", "pdf", "Read and create PDF documents.", skill.Manifest{
			Phrases: []string{"extract a pdf"}, Extensions: []string{".pdf"}, Environments: []string{"documents"},
		}),
		routeFixture("docs:word", "word", "Read and create Word documents.", skill.Manifest{}),
	}}

	result := Rank(Request{
		Query: "Please extract a PDF report", Host: "claude-code", Environment: "documents", Extensions: []string{".pdf"}, Limit: 1,
	}, generation)

	if len(result.Candidates) != 1 || result.Candidates[0].ID != "docs:pdf" {
		t.Fatalf("unexpected route: %#v", result.Candidates)
	}
	for _, reason := range []string{"phrase:extract a pdf", "extension:.pdf", "environment:documents"} {
		assertReason(t, result.Candidates[0], reason)
	}
}

func TestRankEnforcesLimitAndStableIDTieBreak(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("z:last", "last", "Review code.", skill.Manifest{}),
		routeFixture("a:first", "first", "Review code.", skill.Manifest{}),
	}}

	result := Rank(Request{Query: "review code", Limit: 1}, generation)

	if len(result.Candidates) != 1 || result.Candidates[0].ID != "a:first" {
		t.Fatalf("limit or tie-break failed: %#v", result.Candidates)
	}
}

func TestRankReturnsEmptyForNoLexicalMatch(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("docs:pdf", "pdf", "Read PDF documents.", skill.Manifest{}),
	}}

	result := Rank(Request{Query: "what is the weather", Limit: 5}, generation)

	if len(result.Candidates) != 0 {
		t.Fatalf("expected no candidates, got %#v", result.Candidates)
	}
}

func assertReason(t *testing.T, candidate Candidate, want string) {
	t.Helper()
	for _, reason := range candidate.Reasons {
		if reason == want {
			return
		}
	}
	t.Fatalf("reason %q not found in %#v", want, candidate.Reasons)
}
