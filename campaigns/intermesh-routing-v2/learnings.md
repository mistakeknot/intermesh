# Interlab learnings: Intermesh routing V2

## Final summary

- **Starting:** c03-derived top-three recall 0.00; expected verification skill
  ranked fourth
- **Ending:** c03-derived top-three recall 1.00; expected skill ranked third
- **Improvement:** +1.00 absolute
- **Development guardrails:** no-match recall 0.7333, precision 1.00, top-three
  0.9667, top-five 1.00, and MRR 0.8417 were unchanged
- **Locked holdout:** no-match recall 0.40, top-three 0.60, top-five 0.70, and
  MRR 0.4558 were unchanged
- **Experiments:** one mutation after baseline; one retained

The retained implementation is commit `6b201f2e203a`. The sanitized public
regression is derived from trustworthy canary session c03 but is not counted as
another observed route label.

## Validated insights

- A filename such as `AGENTS.md` must not receive the 100-point exact-name
  bonus for a skill named `agents`. Identifier-boundary matching removes this
  false positive while preserving ordinary lexical evidence.
- Live canary failures can seed deterministic public regression cases without
  conflating synthetic evaluation rows with private observed outcomes.

## Dead ends

- A shorter c03 paraphrase already placed verification second, so it could not
  establish the observed failure as the campaign baseline.

## Patterns

- Reproduce the entire routing context before optimizing a failure; seemingly
  incidental operational wording can change the top-three boundary.
- Keep observed-derived regression metrics separate from activation evidence.
