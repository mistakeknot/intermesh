# Interlab learnings: Intermesh routing V4

## Final summary

- **Starting:** observed-derived top-three recall 0.50, arithmetic no-match
  recall/precision 0, and observed MRR 0.50
- **Ending:** observed-derived top-three recall 1.00, arithmetic no-match
  recall/precision 1.00, and observed MRR 0.6667
- **Development guardrails:** no-match recall 0.7333, precision 1.00,
  top-three 0.9667, and top-five 1.00 were unchanged; MRR improved from
  0.8417 to 0.8456
- **Locked holdout:** no-match recall improved from 0.40 to 0.45, top-three
  from 0.60 to 0.70, top-five from 0.70 to 0.75, and MRR from 0.4558 to
  0.5042
- **Experiments:** two retained mutations after baseline, plus two weight
  probes used to find the minimum effective compound-name signal

The sanitized public regressions are derived from trustworthy canary sessions
c07 and c08 but are not counted as additional observed route labels. The two
retained approaches were recorded as Interlab mutations 9 and 10.

## Validated insights

- Function words that survive tokenization can generate arbitrary candidate
  sets even for obvious no-skill requests. `by` was the sole evidence behind
  all three c08 candidates.
- A query-token match inside a compound skill name is stronger evidence than
  the same token appearing only in prose, but it should remain much weaker
  than an exact skill-name invocation.
- A 0.9x name-token IDF signal was the smallest tested weight that put
  `clavain:code-review-discipline` inside c07's top three. It also improved
  rather than degraded the broader development and locked-holdout metrics.

## Dead ends

- A 1.5x signal over-ordered compound-name matches and reduced development
  MRR.
- A 0.8x signal left the expected review skill one position below the bounded
  candidate set.

## Patterns

- Optimize to the bounded candidate-set boundary, not automatically to rank
  one; the agent's metadata applicability gate can perform final selection.
- Keep observed no-match metrics beside observed recall metrics. Progressive
  body loading limits false-positive cost, but a true empty route is still
  cheaper and easier to reason about.
