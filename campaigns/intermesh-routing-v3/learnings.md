# Interlab learnings: Intermesh routing V3

## Final summary

- **Starting:** c04-derived MRR 0.3333; expected verification skill ranked third
- **Ending:** c04-derived MRR 1.00; expected skill ranked first
- **Improvement:** +0.6667 absolute (+200%)
- **Development guardrails:** no-match recall 0.7333, precision 1.00, top-three
  0.9667, top-five 1.00, and MRR 0.8417 were unchanged
- **Locked holdout:** no-match recall 0.40, top-three 0.60, top-five 0.70, and
  MRR 0.4558 were unchanged
- **Experiments:** one mutation after baseline; one retained

The retained implementation is commit `baafc054ca62`. It was recorded as
Interlab mutation 8 with quality signal 1.0.

## Validated insights

- Directly negated action terms are constraints, not positive routing evidence.
  Removing `edit` and `run` after `not` eliminated irrelevant autoresearch and
  document scores without hiding the still-positive `verification commands`.
- Skipping stopwords without ending negation scope correctly handles phrases
  such as `not to edit`, while limiting the behavior to the next content term
  keeps the deterministic parser narrow.

## Dead ends

- None; the first scoped mutation passed all development and holdout gates.

## Patterns

- Candidate-set precision matters even when top-three recall passes because
  every false-positive skill body consumes model context and task latency.
- Use real canary agent behavior—not candidate IDs alone—to decide whether a
  route is correct, partial, or wrong.
