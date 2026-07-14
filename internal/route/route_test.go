package route

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
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

func TestRankIncludesDescriptionForProgressiveSelection(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("intertest:verify", "verify", "Verify completed engineering work.", skill.Manifest{ConflictsWith: []string{"fixture:fast"}}),
	}}

	result := Rank(Request{Query: "verify engineering work", Limit: 1}, generation)
	payload, err := json.Marshal(result.Candidates[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"description":"Verify completed engineering work."`) {
		t.Fatalf("candidate description missing from route JSON: %s", payload)
	}
	if !strings.Contains(string(payload), `"conflicts_with":["fixture:fast"]`) {
		t.Fatalf("candidate conflict metadata missing from route JSON: %s", payload)
	}
}

func TestRankDoesNotTreatFileStemAsExactSkillName(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("intermux:agents", "agents", "Show a dashboard of agent sessions. Not for code quality.", skill.Manifest{}),
		routeFixture("intertest:verification-before-completion", "verification-before-completion", "Use before completion to identify and run verification commands.", skill.Manifest{}),
	}}

	result := Rank(Request{
		Query: "inspect AGENTS.md and identify the verification commands required before completion",
		Limit: 3,
	}, generation)

	if len(result.Candidates) == 0 || result.Candidates[0].ID != "intertest:verification-before-completion" {
		t.Fatalf("file stem received an exact-name bonus: %#v", result.Candidates)
	}
	for _, candidate := range result.Candidates {
		if candidate.ID == "intermux:agents" && candidate.Components["exact_name"] != 0 {
			t.Fatalf("AGENTS.md must not be an exact agents invocation: %#v", candidate)
		}
	}
}

func TestRankDoesNotTreatNegatedActionsAsPositiveEvidence(t *testing.T) {
	generation := registry.Generation{Skills: []skill.Skill{
		routeFixture("interlab:autoresearch", "autoresearch", "Run a continuous optimization loop — edit code, benchmark, keep or discard, repeat. Use when systematically optimizing a metric through iterative code changes.", skill.Manifest{}),
		routeFixture("documents:documents", "documents", "Create, edit, redline, and comment on DOCX artifacts with a strict render-and-verify workflow, then iterate before delivering the final document.", skill.Manifest{}),
		routeFixture("intertest:verification-before-completion", "verification-before-completion", "Use before completion to identify and run verification commands.", skill.Manifest{}),
	}}

	result := Rank(Request{
		Query: "Route this request through Intermesh, then inspect AGENTS.md and identify the verification commands required before completion. Do not edit project files and do not run the verification commands.",
		Limit: 3,
	}, generation)

	if len(result.Candidates) == 0 || result.Candidates[0].ID != "intertest:verification-before-completion" {
		t.Fatalf("negated actions outranked verification: %#v", result.Candidates)
	}
	for _, candidate := range result.Candidates {
		if candidate.Components["lexical:edit"] != 0 || candidate.Components["lexical:run"] != 0 {
			t.Fatalf("negated terms contributed lexical evidence: %#v", candidate)
		}
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

func TestTokensDropConversationalFunctionWords(t *testing.T) {
	got := tokens("what should you give me after this than before no do did does your good plus about how make are have weather forecast")
	want := []string{"before", "no", "weather", "forecast"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens=%#v want %#v", got, want)
	}
}

func TestQueryTokensDropDirectlyNegatedTerms(t *testing.T) {
	got := queryTokens("Do not edit project files and do not run verification commands without deploying releases")
	want := []string{"project", "files", "verification", "commands", "releases"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("queryTokens=%#v want %#v", got, want)
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
