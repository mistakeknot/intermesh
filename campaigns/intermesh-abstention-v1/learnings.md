# Interlab learnings: Intermesh abstention V1

## Final summary

- **Development starting:** no-match recall 0.20
- **Development retained:** no-match recall 0.7333
- **Development improvement:** +0.5333 absolute (+266.7%)
- **Development guardrails:** precision 1.00, top-3 0.9667, top-5 1.00,
  and MRR 0.8417 were unchanged
- **Holdout starting:** no-match recall 0.05, top-3 0.60, top-5 0.70
- **Holdout retained:** no-match recall 0.40, top-3 0.60, top-5 0.70
- **Experiments:** four mutations after baseline; two retained, one rejected by
  the development gate, one reverted after the locked holdout

The retained implementation is commits `b9adc248b0f0` and `10d357d79015`.
Commit `2d867b2809fb` reached perfect development abstention but was reverted by
`d5f18fd` after it reduced holdout top-3/top-5 to 0.50/0.60 and holdout MRR by
more than 20%.

## Validated insights

- Conversational function words were responsible for most lexical false
  positives. Removing them improved development and holdout no-match recall
  without changing positive recall.
- Route-level evidence breadth can perfectly separate the development set, but
  the direct positive corpus does not represent indirect paraphrases well
  enough to validate that gate.
- A locked paraphrase holdout is essential. Development guardrails alone would
  have accepted a mutation that materially reduced generalization.

## Dead ends

- Filtering every single-overlap candidate individually reached 1.00 no-match
  recall but dropped development top-3 below 0.95.
- Abstaining when every candidate had one weak overlap passed all development
  gates but regressed holdout positive recall, so it was reverted.
- More lexical thresholds, winning margins, and coverage heuristics are deferred
  until the positive development set includes indirect paraphrases and the
  campaign metric represents both abstention and positive generalization.

## Patterns

- Optimize abstention with a balanced or constrained objective across direct
  and paraphrased positive cases; primary no-match recall alone is incomplete.
- Keep synthetic development and holdout evidence separate from trustworthy
  observed outcomes. Neither lifts the real-profile activation hold.
- Interlab mutation provenance currently ranks primary quality without knowing
  the keep/crash decision. A constraint-violating mutation with primary 1.00
  became the store's `best_quality`; consumers must inspect decision metadata
  until Interlab makes validity part of its ranking key.
