# Testing and hill climbing

## Safety position

Intermesh V0 remains opt-in and routing-only by default. Synthetic benchmark
improvements do not satisfy the activation gate: a reduced real host catalog
still requires at least 30 trustworthy observed route outcomes and improved
abstention behavior. Use `profile plan` against real catalogs; do not run
`profile apply` as part of an optimization campaign.

## Install and index

From a clean Intermesh checkout:

```bash
mkdir -p "$HOME/.local/bin"
GOBIN="$HOME/.local/bin" go install ./cmd/intermesh
export PATH="$HOME/.local/bin:$PATH"
scripts/index-codex-catalog.sh
intermesh doctor
```

Override the derived registry location when isolating a trial:

```bash
scripts/index-codex-catalog.sh --db /tmp/intermesh-registry.db
intermesh doctor --db /tmp/intermesh-registry.db
```

The index helper scans:

- `~/.codex/<plugin>/skills/<skill>/SKILL.md`
- `~/.codex/plugins/cache/<provider>/<plugin>/<version>/skills/<skill>/SKILL.md`
- `~/projects/dotfiles/common/.codex/skills`, using separate `local` and
  `codex-system` namespaces

It passes explicit skill roots to the CLI and never executes skill code.

## Verification layers

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/intermesh ./cmd/intermesh-eval
bash tests/shell/test_setup_scripts.sh
bash tests/shell/test_codex_canary.sh
bash scripts/experiment.sh --check-gates
bash scripts/interlab-abstention.sh
```

The abstention benchmark rebuilds a temporary registry, runs route/eval tests,
evaluates the 60-case development corpus, and emits:

- primary `no_match_recall`
- secondary `no_match_precision`, `top3_recall`, `top5_recall`, `mrr`, and
  `warm_p95_micros`

It exits nonzero if top-3 or top-5 recall falls below 95%, no-match precision
falls below 80%, or warm p95 reaches 50 ms.

## Corpus policy

`testdata/routes/abstention-dev.jsonl` contains 30 positive and 30 no-match
synthetic cases. `abstention-holdout.jsonl` contains a disjoint 20+20
paraphrase set. Both are labeled synthetic and must not be reported as observed
production evidence.

The 2026-07-14 starting measurements against 170 current Codex skill roots are:

| Corpus | No-match recall | No-match precision | Top-3 | Top-5 | MRR |
|---|---:|---:|---:|---:|---:|
| Development | 20% | 100% | 96.7% | 100% | 0.842 |
| Locked holdout | 5% | 100% | 60% | 70% | 0.456 |

The holdout is evaluated before and after a campaign, not used to choose each
mutation. Its weak starting positive recall is direct evidence that the current
lexical router does not yet generalize well to indirect paraphrases.

Report holdout metrics without applying the development guardrails:

```bash
bash scripts/interlab-abstention.sh \
  --cases testdata/routes/abstention-holdout.jsonl \
  --report-only
```

## First campaign result

The `intermesh-abstention-v1` campaign retained two stopword mutations and
reverted an over-aggressive route-level gate after the holdout exposed a
generalization regression:

| Corpus | No-match before | No-match retained | Positive top-3/top-5 |
|---|---:|---:|---:|
| Original V0 four-case set | 25% | 100% | Adversarial match cases unchanged at 100%/100% |
| Development | 20% | 73.3% | Unchanged at 96.7%/100% |
| Locked holdout | 5% | 40% | Unchanged at 60%/70% |

The development-perfect route-level gate was not retained: it reduced holdout
top-3/top-5 to 50%/60% and MRR by more than 20%. The raw campaign record and
validated learnings are archived under `campaigns/intermesh-abstention-v1/`.

## Register Interlab with Codex

The Interlab skill alone is not enough; Codex must also expose its stateless MCP
experiment tools:

```bash
cd "$HOME/.codex/interlab"
mkdir -p "$HOME/.local/bin"
GOBIN="$HOME/.local/bin" go install ./cmd/interlab-mcp
codex mcp add interlab -- "$HOME/.local/bin/interlab-mcp"
codex mcp list
```

Restart Codex after registration. For an on-demand session, the same server can
be driven with `~/.local/bin/mcp` without preloading its schemas.

## Second campaign result

Canary session c03 exposed a filename collision: `AGENTS.md` gave the generic
`intermux:agents` skill a 100-point exact-name bonus, pushing the expected
verification skill to rank four. The `intermesh-routing-v2` campaign retained
identifier-boundary exact-name matching:

| Metric | Before | Retained |
|---|---:|---:|
| c03-derived top-3 recall | 0.00 | 1.00 |
| Development no-match recall | 0.7333 | 0.7333 |
| Development top-3/top-5 | 0.9667 / 1.00 | 0.9667 / 1.00 |
| Locked holdout top-3/top-5 | 0.60 / 0.70 | 0.60 / 0.70 |

The public regression is observed-derived, not an additional trustworthy
canary label. Campaign evidence is archived under
`campaigns/intermesh-routing-v2/`.

Canary session c04 then showed that negated constraints still acted as positive
evidence: `do not edit` and `do not run` selected autoresearch and documents
ahead of verification. The `intermesh-routing-v3` campaign excluded the next
content term after `not`, `no`, `never`, or `without` from positive lexical
scoring. The c04-derived expected skill moved from rank three to rank one (MRR
0.3333 to 1.00), with development and locked-holdout metrics unchanged. Its
evidence is archived under `campaigns/intermesh-routing-v3/`.

## Campaign contract

Initialize an `intermesh-abstention` campaign with:

- metric: `no_match_recall`, unit `ratio`, direction `higher_is_better`
- benchmark: `bash scripts/interlab-abstention.sh`
- mutable files: `internal/route/route.go`, `internal/route/score.go`,
  `internal/route/tokenize.go`, and `internal/route/route_test.go`
- immutable constraints: no new runtime dependency; deterministic results;
  positive recall/precision/latency gates remain green; registry and receipt
  contracts unchanged; no profile mutation

Each experiment changes one routing behavior. Run the route tests before the
benchmark. Keep a mutation only when the primary improves and secondaries stay
within their gates; otherwise discard it. Interspect remains the owner of
learned production priors and outcome calibration.

## Production feedback loop

Default routes append privacy-minimal `intermesh.route.v1` records to
`${XDG_STATE_HOME:-$HOME/.local/state}/intermesh/routes.jsonl`. Interspect can
ingest these as decision evidence, but a route is not a success signal. The
activation hold remains until host/session outcome evidence is joined to route
IDs and at least 30 trustworthy labels can be evaluated.

Use the [Codex router-only canary](codex-router-canary.md) to collect those
labels without changing the normal profile. Twenty labeled sessions produce an
interim report; the activation gate remains closed until all 30 sessions are
labeled and every context, routing, coverage, and task-outcome threshold passes.
