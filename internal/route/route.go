package route

import (
	"sort"

	"github.com/mistakeknot/intermesh/internal/registry"
)

const DefaultLimit = 5

type Request struct {
	Query       string
	Host        string
	CWD         string
	Environment string
	Extensions  []string
	Limit       int
}

type Candidate struct {
	ID          string             `json:"id"`
	Description string             `json:"description,omitempty"`
	SkillMD     string             `json:"skill_md"`
	Score       float64            `json:"score"`
	Reasons     []string           `json:"reasons"`
	Components  map[string]float64 `json:"components,omitempty"`
	SelectedBy  string             `json:"selected_by"`
	RequiredBy  []string           `json:"required_by"`
}

type Result struct {
	Version             int         `json:"version"`
	Query               string      `json:"query,omitempty"`
	Host                string      `json:"host,omitempty"`
	RegistryFingerprint string      `json:"registry_fingerprint"`
	Candidates          []Candidate `json:"candidates"`
	Warnings            []string    `json:"warnings"`
}

func Rank(request Request, generation registry.Generation) Result {
	limit := request.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	scoring := newScorer(generation.Skills)
	result := Result{
		Version: 1, Query: request.Query, Host: request.Host,
		RegistryFingerprint: generation.Fingerprint,
		Candidates:          []Candidate{}, Warnings: []string{},
	}
	for _, item := range generation.Skills {
		score, reasons, components := scoring.score(request, item)
		if score <= 0 {
			continue
		}
		result.Candidates = append(result.Candidates, Candidate{
			ID: item.ID, Description: item.Description, SkillMD: item.SkillMD, Score: score, Reasons: reasons, Components: components,
			SelectedBy: "rank", RequiredBy: []string{},
		})
	}
	sort.Slice(result.Candidates, func(i, j int) bool {
		if result.Candidates[i].Score == result.Candidates[j].Score {
			return result.Candidates[i].ID < result.Candidates[j].ID
		}
		return result.Candidates[i].Score > result.Candidates[j].Score
	})
	if len(result.Candidates) > limit {
		result.Candidates = result.Candidates[:limit]
	}
	return result
}
