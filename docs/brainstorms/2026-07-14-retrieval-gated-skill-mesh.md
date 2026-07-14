---
artifact_type: brainstorm
bead: sylveste-qn2c
date: 2026-07-14
status: validated
---
# Retrieval-gated skill mesh

## Decision

Build Intermesh as a host-agnostic CLI and derived local registry, then keep each agent integration thin. The experiment must demonstrate that candidate retrieval preserves relevant-skill recall while reducing metadata exposed to the model. It must also prove that the host can actually isolate unselected skills; emitting a hint into an already-bloated prompt does not count.

The current explicit project decision assigns `Intermesh` to this skill-routing substrate. This supersedes the unused 2026-04-29 reservation of the same name for a possible future coordination substrate. Coordination remains owned by Intermute, Intermux, Interlock, and Athenmesh; if a new substrate is eventually justified, it needs a different name.

**Alignment:** This turns a large prompt-time capability list into a small, explicit Unix-style retrieval boundary with measurable receipts.

**Conflict/Risk:** The chosen name conflicts with an older reservation, and host catalog isolation varies substantially across Codex, Claude Code, and Hermes. Both must be resolved explicitly rather than hidden behind a generic adapter claim.

## Experiment outcome

V0 passed its deterministic recall, metadata-reduction, warm-latency, and non-mutation gates. The operational adapters request three candidates because top-3 deterministic recall remained 96.7% while loading five full skill bodies produced an unacceptable 79 KB p95 context tail. Observed recall remains insufficient (zero trustworthy query→skill labels), and no-match recall is only 25%, so real catalog activation is opt-in rather than automatic. See `docs/reports/intermesh-v0-experiment.md`.

## Outcome

A request such as `review this pull request for security regressions` should produce a compact route result containing the best few skills, reasons, paths, and declared dependencies. The host then reads only those `SKILL.md` files. The full catalog remains on disk and consumes no prompt tokens.

## Ownership

| Concern | Owner |
|---|---|
| Canonical instructions and triggers | Each skill's `SKILL.md` |
| Declarative requires/conflicts/composition | Adjacent `intermesh.yaml` |
| Discovery, indexing, search, resolution | Intermesh |
| Skill creation and audit | Interskill |
| Outcome evidence and learned calibration | Interspect |
| Run, phase, gate, and dispatch state | Intercore |
| Host-specific catalog activation | Thin Intermesh adapter/profile |

## Options considered

### A. Keep every description loaded and shorten them

This is safe and immediately available, but savings plateau because every installed skill still pays a fixed discovery cost. The April audit found 32,330 description bytes across 156 skills even after deduplication.

### B. Add a router hint hook

Already tested by `sylveste-a4oj.9.1`. It can reduce deliberation for explicit commands but cannot remove host-provided skill metadata. It may add context rather than save it.

### C. Retrieval-gated external catalog — selected

Keep only a small router adapter in the host's automatic catalog. Index the larger library out of band, retrieve top-k candidates, resolve relationships, and load selected files. This has the largest potential saving but requires safe profile isolation and per-host proof.

### D. Put the graph in Intercore

Rejected. Skill discovery is derived capability metadata, not orchestration state. Coupling it to Intercore would broaden the kernel and make standalone adoption harder.

## V0 mechanism

1. Scan configured roots without following unsafe escape paths.
2. Parse frontmatter and optional relationship manifests without executing skill content.
3. Replace the registry transactionally so a failed rebuild leaves the last good index intact.
4. Rank with deterministic exact-name, trigger, environment, path/filetype, lexical, and usage-prior signals.
5. Resolve `requires`, then report rather than silently discard conflicts.
6. Return compact versioned JSON and append an outcome-neutral route receipt.
7. Let Interspect later attach actual outcomes and calibrated boosts.

Semantic reranking may be added through Intersearch after the lexical baseline is measured. It is not required to prove the context architecture.

## Experiment gates

- Index every valid skill in the current Codex catalog and report invalid entries without aborting the rebuild.
- Achieve top-5 recall of at least 95% on deterministic explicit/filetype cases.
- Achieve top-5 recall of at least 85% on an observed invocation corpus where trustworthy labels exist.
- Reduce modeled discovery metadata by at least 80% for router plus top-5 candidates versus the full current catalog.
- Keep warm p95 routing latency below 50 ms on the local catalog.
- Produce identical index fingerprints from identical source files.
- Demonstrate profile dry-run and lossless restore without modifying the real catalog during tests.

## Kill or rescope conditions

- If a host cannot isolate unselected skill metadata without unsupported configuration mutation, document that adapter as routing-only rather than claiming context savings.
- If observed top-5 recall is below 85%, do not activate the reduced catalog by default; improve retrieval or retain a larger always-on tier.
- If catalog activation cannot be made reversible and crash-safe, ship search/diagnostics only.
