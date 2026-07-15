# Interlab learnings: Intermesh routing V5

## Final summary

- **Starting:** observed no-match recall 0.20 and development no-match recall
  0.7333
- **Ending:** observed no-match recall 1.00 and development no-match recall
  0.9333
- **Development guardrails:** no-match precision remained 1.00, top-three
  improved from 0.9667 to 1.00, top-five remained 1.00, and MRR improved from
  0.8456 to 0.8722
- **Locked holdout:** no-match recall improved from 0.45 to 0.95, precision
  remained 1.00, top-three improved from 0.70 to 0.80, top-five from 0.75 to
  0.80, and MRR from 0.5042 to 0.6000
- **Experiments:** one retained mutation after baseline, with two refinement
  probes used to restore positive recall

The sanitized public regressions come from trustworthy canary sessions c11,
c12, c14, c15, and c16 but are not counted as additional observed outcomes.
The retained approach was recorded as Interlab mutation 11.

## Validated insights

- A single loose lexical overlap is not enough evidence to spend a bounded
  candidate slot. The observed false positives were supported only by
  `files`, `write`, `20`, or `convert`.
- Exact and declarative signals remain sufficient alone. Weak inferred signals
  require two independent components.
- Simple singular aliases recover legitimate evidence hidden by morphology:
  `plans` matches `plan`, and `illustrations` matches `illustration`.

## Dead ends

- Applying the two-component rule without morphology support dropped
  `clavain:writing-plans` from the development top five and made the raster
  image holdout an empty route.

## Patterns

- Candidate admission and candidate ordering are separate optimization
  surfaces. Filter weak evidence before ranking rather than forcing the agent
  to reject cheap but irrelevant candidates.
- Every abstention mutation must be checked against both positive recall and a
  locked holdout; development no-match gains alone are not enough.
