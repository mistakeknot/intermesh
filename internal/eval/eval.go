package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
)

type Case struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind,omitempty"`
	Query       string   `json:"query"`
	Host        string   `json:"host,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Extensions  []string `json:"extensions,omitempty"`
	Expected    []string `json:"expected,omitempty"`
	NoMatch     bool     `json:"no_match,omitempty"`
}

type Metrics struct {
	Cases            int     `json:"cases"`
	MatchCases       int     `json:"match_cases"`
	NoMatchCases     int     `json:"no_match_cases"`
	PredictedNoMatch int     `json:"predicted_no_match"`
	Top1Recall       float64 `json:"top_1_recall"`
	Top3Recall       float64 `json:"top_3_recall"`
	Top5Recall       float64 `json:"top_5_recall"`
	MRR              float64 `json:"mrr"`
	NoMatchPrecision float64 `json:"no_match_precision"`
	NoMatchRecall    float64 `json:"no_match_recall"`
}

func LoadCases(reader io.Reader) ([]Case, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	seen := map[string]struct{}{}
	var cases []Case
	for line := 1; scanner.Scan(); line++ {
		content := strings.TrimSpace(scanner.Text())
		if content == "" {
			continue
		}
		var item Case
		if err := json.Unmarshal([]byte(content), &item); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		item.ID = strings.TrimSpace(item.ID)
		item.Query = strings.TrimSpace(item.Query)
		if item.ID == "" || item.Query == "" {
			return nil, fmt.Errorf("line %d: id and query are required", line)
		}
		if _, exists := seen[item.ID]; exists {
			return nil, fmt.Errorf("line %d: duplicate case id %q", line, item.ID)
		}
		if item.NoMatch == (len(item.Expected) > 0) {
			return nil, fmt.Errorf("line %d: exactly one of expected or no_match is required", line)
		}
		seen[item.ID] = struct{}{}
		cases = append(cases, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cases, nil
}

func Score(cases []Case, predictions map[string][]string) Metrics {
	metrics := Metrics{Cases: len(cases)}
	var reciprocalRank float64
	var correctNoMatch int
	for _, item := range cases {
		predicted := predictions[item.ID]
		if len(predicted) == 0 {
			metrics.PredictedNoMatch++
			if item.NoMatch {
				correctNoMatch++
			}
		}
		if item.NoMatch {
			metrics.NoMatchCases++
			continue
		}
		metrics.MatchCases++
		expected := make(map[string]struct{}, len(item.Expected))
		for _, id := range item.Expected {
			expected[id] = struct{}{}
		}
		firstRank := 0
		for index, id := range predicted {
			if _, ok := expected[id]; ok {
				firstRank = index + 1
				break
			}
		}
		if firstRank == 1 {
			metrics.Top1Recall++
		}
		if firstRank > 0 && firstRank <= 3 {
			metrics.Top3Recall++
		}
		if firstRank > 0 && firstRank <= 5 {
			metrics.Top5Recall++
		}
		if firstRank > 0 {
			reciprocalRank += 1 / float64(firstRank)
		}
	}
	if metrics.MatchCases > 0 {
		denominator := float64(metrics.MatchCases)
		metrics.Top1Recall /= denominator
		metrics.Top3Recall /= denominator
		metrics.Top5Recall /= denominator
		metrics.MRR = reciprocalRank / denominator
	}
	if metrics.PredictedNoMatch > 0 {
		metrics.NoMatchPrecision = float64(correctNoMatch) / float64(metrics.PredictedNoMatch)
	}
	if metrics.NoMatchCases > 0 {
		metrics.NoMatchRecall = float64(correctNoMatch) / float64(metrics.NoMatchCases)
	}
	return metrics
}

func Percentile(values []int64, quantile float64) int64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]int64(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	if quantile <= 0 {
		return ordered[0]
	}
	if quantile > 1 {
		quantile = 1
	}
	rank := int(math.Ceil(quantile*float64(len(ordered)))) - 1
	return ordered[rank]
}
