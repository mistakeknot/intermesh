package route

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/mistakeknot/intermesh/internal/skill"
)

type scorer struct {
	documentFrequency map[string]int
	documentCount     int
}

func newScorer(skills []skill.Skill) scorer {
	documentFrequency := make(map[string]int)
	for _, item := range skills {
		terms := make(map[string]struct{})
		for _, term := range tokens(searchableText(item)) {
			terms[term] = struct{}{}
		}
		for term := range terms {
			documentFrequency[term]++
		}
	}
	return scorer{documentFrequency: documentFrequency, documentCount: len(skills)}
}

func (s scorer) score(request Request, item skill.Skill) (float64, []string, map[string]float64) {
	query := normalizeText(request.Query)
	if query == "" {
		return 0, nil, nil
	}
	if len(item.Manifest.Environments) > 0 && request.Environment != "" && !containsFold(item.Manifest.Environments, request.Environment) {
		return 0, nil, nil
	}

	var score float64
	var reasons []string
	components := make(map[string]float64)
	add := func(key string, value float64) {
		score += value
		components[key] += value
	}
	lowerID := strings.ToLower(item.ID)
	lowerName := strings.ToLower(item.Name)
	if containsIdentifier(query, lowerID) {
		add("exact_id", 100)
		reasons = append(reasons, "exact_id")
	} else if containsIdentifier(query, lowerName) {
		add("exact_name", 100)
		reasons = append(reasons, "exact_name")
	}
	if strings.Contains(query, "/"+lowerID) {
		add("namespaced_command", 80)
		reasons = append(reasons, "namespaced_command")
	}
	for _, phrase := range item.Manifest.Phrases {
		normalizedPhrase := normalizeText(phrase)
		if normalizedPhrase != "" && strings.Contains(query, normalizedPhrase) {
			key := "phrase:" + normalizedPhrase
			add(key, 40)
			reasons = append(reasons, key)
		}
	}
	for _, extension := range request.Extensions {
		if containsFold(item.Manifest.Extensions, extension) {
			key := "extension:" + strings.ToLower(extension)
			add(key, 20)
			reasons = append(reasons, key)
		}
	}
	if request.Environment != "" && containsFold(item.Manifest.Environments, request.Environment) {
		key := "environment:" + strings.ToLower(request.Environment)
		add(key, 10)
		reasons = append(reasons, key)
	}

	documentTerms := make(map[string]struct{})
	for _, term := range tokens(searchableText(item)) {
		documentTerms[term] = struct{}{}
	}
	for _, term := range queryTokens(query) {
		if _, exists := documentTerms[term]; !exists {
			continue
		}
		idf := math.Log(float64(s.documentCount+1)/float64(s.documentFrequency[term]+1)) + 1
		key := fmt.Sprintf("lexical:%s", term)
		add(key, idf)
		reasons = append(reasons, key)
	}
	sort.Strings(reasons)
	return score, reasons, components
}

func searchableText(item skill.Skill) string {
	return strings.Join(append([]string{item.ID, item.Name, item.Description}, item.Manifest.Phrases...), " ")
}

func containsFold(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(expected)) {
			return true
		}
	}
	return false
}

func containsIdentifier(query, identifier string) bool {
	for _, field := range strings.Fields(query) {
		trimmed := strings.Trim(field, "`'\".,;()[]{}")
		if trimmed == identifier || strings.TrimPrefix(trimmed, "/") == identifier {
			return true
		}
	}
	return false
}
