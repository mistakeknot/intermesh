package eval

import (
	"os"
	"strings"
	"testing"
)

func TestScoreComputesRecallMRRAndNoMatchPrecision(t *testing.T) {
	cases := []Case{
		{ID: "first", Expected: []string{"skill:a"}},
		{ID: "second", Expected: []string{"skill:b", "skill:c"}},
		{ID: "miss", Expected: []string{"skill:d"}},
		{ID: "none-correct", NoMatch: true},
		{ID: "none-wrong", NoMatch: true},
	}
	predictions := map[string][]string{
		"first":        {"skill:a", "skill:x"},
		"second":       {"skill:x", "skill:c"},
		"miss":         {"skill:x"},
		"none-correct": {},
		"none-wrong":   {"skill:x"},
	}

	metrics := Score(cases, predictions)

	if metrics.MatchCases != 3 || metrics.NoMatchCases != 2 {
		t.Fatalf("unexpected denominators: %#v", metrics)
	}
	assertNear(t, metrics.Top1Recall, 1.0/3.0)
	assertNear(t, metrics.Top3Recall, 2.0/3.0)
	assertNear(t, metrics.Top5Recall, 2.0/3.0)
	assertNear(t, metrics.MRR, 0.5)
	assertNear(t, metrics.NoMatchPrecision, 1.0)
	assertNear(t, metrics.NoMatchRecall, 0.5)
	if metrics.PredictedNoMatch != 1 {
		t.Fatalf("unexpected no-match prediction count: %#v", metrics)
	}
}

func TestLoadCasesRejectsDuplicateIDsAndInvalidLabels(t *testing.T) {
	_, err := LoadCases(strings.NewReader("{\"id\":\"same\",\"query\":\"q\",\"expected\":[\"a\"]}\n{\"id\":\"same\",\"query\":\"q2\",\"no_match\":true}\n"))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}

	_, err = LoadCases(strings.NewReader("{\"id\":\"bad\",\"query\":\"q\",\"expected\":[\"a\"],\"no_match\":true}\n"))
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("expected label rejection, got %v", err)
	}
}

func TestPercentileUsesNearestRank(t *testing.T) {
	values := []int64{50, 10, 40, 20, 30}
	if got := Percentile(values, 0.50); got != 30 {
		t.Fatalf("p50=%d", got)
	}
	if got := Percentile(values, 0.95); got != 50 {
		t.Fatalf("p95=%d", got)
	}
}

func TestAbstentionCorporaAreSizedAndDisjoint(t *testing.T) {
	development := loadCorpusFile(t, "../../testdata/routes/abstention-dev.jsonl")
	holdout := loadCorpusFile(t, "../../testdata/routes/abstention-holdout.jsonl")

	assertCorpusBalance(t, "development", development, 30, 30)
	assertCorpusBalance(t, "holdout", holdout, 20, 20)

	developmentIDs := make(map[string]struct{}, len(development))
	for _, item := range development {
		developmentIDs[item.ID] = struct{}{}
	}
	for _, item := range holdout {
		if _, exists := developmentIDs[item.ID]; exists {
			t.Fatalf("case %q appears in both development and holdout corpora", item.ID)
		}
	}
}

func loadCorpusFile(t *testing.T, path string) []Case {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cases, err := LoadCases(file)
	if err != nil {
		t.Fatal(err)
	}
	return cases
}

func assertCorpusBalance(t *testing.T, name string, cases []Case, minimumMatch, minimumNoMatch int) {
	t.Helper()
	var match, noMatch int
	for _, item := range cases {
		if item.NoMatch {
			noMatch++
		} else {
			match++
		}
	}
	if match < minimumMatch || noMatch < minimumNoMatch {
		t.Fatalf("%s corpus has %d match and %d no-match cases; need at least %d and %d", name, match, noMatch, minimumMatch, minimumNoMatch)
	}
}

func assertNear(t *testing.T, got, want float64) {
	t.Helper()
	if got < want-1e-9 || got > want+1e-9 {
		t.Fatalf("got %.12f want %.12f", got, want)
	}
}
