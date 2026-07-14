package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mistakeknot/intermesh/internal/registry"
	"github.com/mistakeknot/intermesh/internal/route"
	"github.com/mistakeknot/intermesh/internal/skill"
)

func Resolve(input route.Result, generation registry.Generation) (route.Result, error) {
	byID := make(map[string]skill.Skill, len(generation.Skills))
	for _, item := range generation.Skills {
		byID[item.ID] = item
	}
	original := make(map[string]route.Candidate, len(input.Candidates))
	for _, candidate := range input.Candidates {
		original[candidate.ID] = candidate
	}

	state := make(map[string]uint8)
	requiredBy := make(map[string]map[string]struct{})
	var ordered []string
	var visit func(string, []string) error
	visit = func(id string, stack []string) error {
		item, exists := byID[id]
		if !exists {
			return fmt.Errorf("required skill %q is missing from the registry", id)
		}
		if state[id] == 2 {
			return nil
		}
		if state[id] == 1 {
			start := 0
			for index, stacked := range stack {
				if stacked == id {
					start = index
					break
				}
			}
			cycle := append(append([]string(nil), stack[start:]...), id)
			return fmt.Errorf("requirement cycle: %s", strings.Join(cycle, " -> "))
		}
		state[id] = 1
		stack = append(stack, id)
		for _, requirement := range item.Manifest.Requires {
			if requiredBy[requirement] == nil {
				requiredBy[requirement] = make(map[string]struct{})
			}
			requiredBy[requirement][id] = struct{}{}
			if err := visit(requirement, stack); err != nil {
				return err
			}
		}
		state[id] = 2
		ordered = append(ordered, id)
		return nil
	}
	for _, candidate := range input.Candidates {
		if err := visit(candidate.ID, nil); err != nil {
			return route.Result{}, err
		}
	}

	output := input
	output.Candidates = make([]route.Candidate, 0, len(ordered))
	selected := make(map[string]struct{}, len(ordered))
	for _, id := range ordered {
		selected[id] = struct{}{}
		candidate, exists := original[id]
		if !exists {
			item := byID[id]
			candidate = route.Candidate{ID: id, Description: item.Description, SkillMD: item.SkillMD, SelectedBy: "requirement", Reasons: []string{}, RequiredBy: []string{}, ConflictsWith: append([]string(nil), item.Manifest.ConflictsWith...)}
		} else {
			item := byID[id]
			if candidate.Description == "" {
				candidate.Description = item.Description
			}
			if candidate.SkillMD == "" {
				candidate.SkillMD = item.SkillMD
			}
			candidate.ConflictsWith = append([]string(nil), item.Manifest.ConflictsWith...)
		}
		sort.Strings(candidate.ConflictsWith)
		parents := requiredBy[id]
		for parent := range parents {
			candidate.RequiredBy = append(candidate.RequiredBy, parent)
		}
		sort.Strings(candidate.RequiredBy)
		output.Candidates = append(output.Candidates, candidate)
	}

	pairs := make(map[string]struct{})
	for id := range selected {
		for _, conflict := range byID[id].Manifest.ConflictsWith {
			if _, exists := selected[conflict]; !exists {
				continue
			}
			left, right := id, conflict
			if right < left {
				left, right = right, left
			}
			key := left + "\x00" + right
			if _, exists := pairs[key]; exists {
				continue
			}
			pairs[key] = struct{}{}
			output.Warnings = append(output.Warnings, fmt.Sprintf("skill %s conflicts with %s", left, right))
		}
	}
	sort.Strings(output.Warnings)
	return output, nil
}
