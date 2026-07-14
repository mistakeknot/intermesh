package eval

import (
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

func assertNear(t *testing.T, got, want float64) {
	t.Helper()
	if got < want-1e-9 || got > want+1e-9 {
		t.Fatalf("got %.12f want %.12f", got, want)
	}
}
