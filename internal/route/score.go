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

func (s scorer) score(request Request, item skill.Skill) (float64, []string) {
	query := normalizeText(request.Query)
	if query == "" {
		return 0, nil
	}
	if len(item.Manifest.Environments) > 0 && request.Environment != "" && !containsFold(item.Manifest.Environments, request.Environment) {
		return 0, nil
	}

	var score float64
	var reasons []string
	lowerID := strings.ToLower(item.ID)
	lowerName := strings.ToLower(item.Name)
	if containsIdentifier(query, lowerID) {
		score += 100
		reasons = append(reasons, "exact_id")
	} else if containsToken(tokens(query), lowerName) {
		score += 100
		reasons = append(reasons, "exact_name")
	}
	if strings.Contains(query, "/"+lowerID) {
		score += 80
		reasons = append(reasons, "namespaced_command")
	}
	for _, phrase := range item.Manifest.Phrases {
		normalizedPhrase := normalizeText(phrase)
		if normalizedPhrase != "" && strings.Contains(query, normalizedPhrase) {
			score += 40
			reasons = append(reasons, "phrase:"+normalizedPhrase)
		}
	}
	for _, extension := range request.Extensions {
		if containsFold(item.Manifest.Extensions, extension) {
			score += 20
			reasons = append(reasons, "extension:"+strings.ToLower(extension))
		}
	}
	if request.Environment != "" && containsFold(item.Manifest.Environments, request.Environment) {
		score += 10
		reasons = append(reasons, "environment:"+strings.ToLower(request.Environment))
	}

	documentTerms := make(map[string]struct{})
	for _, term := range tokens(searchableText(item)) {
		documentTerms[term] = struct{}{}
	}
	for _, term := range tokens(query) {
		if _, exists := documentTerms[term]; !exists {
			continue
		}
		idf := math.Log(float64(s.documentCount+1)/float64(s.documentFrequency[term]+1)) + 1
		score += idf
		reasons = append(reasons, fmt.Sprintf("lexical:%s", term))
	}
	sort.Strings(reasons)
	return score, reasons
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

func containsToken(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
