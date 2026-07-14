---
artifact_type: experiment-report
bead: sylveste-qn2c
date: 2026-07-14
status: passed-with-activation-hold
---
# Intermesh V0 Experiment

## Verdict

Land the CLI, registry, adapters, diagnostics, and reversible profile mechanism. Do **not** activate a reduced real catalog by default yet: deterministic routing passes, but there are no trustworthy observed query→skill labels and the adversarial no-match recall is only 25%.

The experiment proves the architecture can remove always-on skill metadata. It does not prove every routed request uses less total context, because selected `SKILL.md` bodies are larger than their descriptions. The adapters therefore request three candidates, not five.

## Method

- Host and date: Codex on macOS, 2026-07-14.
- Baseline: Sylveste `scripts/perf/audit-skill-contributions.py --plugins-root ~/.codex`.
- Index: current canonical Interverse, local, system, bundled, curated, and primary-runtime skill roots; read-only.
- Corpus: 30 deterministic cases and 10 adversarial cases in `testdata/routes/v0.jsonl`.
- Timing: 800 warm in-process routes after two warmups per case; 50 cold CLI-process routes.
- Safety: `intermesh profile plan` only. No profile apply/restore command touched the real catalog.

The baseline audit counted 180 catalog entries and 32,071 description bytes. Intermesh indexed 170 unique canonical skills and 32,156 description bytes with zero diagnostics. The count difference is caused by the audit and registry using different deduplication identities; the byte totals show they cover the same magnitude of catalog.

## Results

| Metric | Result | Gate |
|---|---:|---:|
| Deterministic top-1 recall | 73.3% | Informational |
| Deterministic top-3 recall | 96.7% | Informational operational limit |
| Deterministic top-5 recall | 100% | ≥95% — pass |
| Deterministic MRR | 0.842 | Informational |
| Adversarial top-3 / top-5 recall | 100% / 100% | Informational |
| Adversarial no-match precision | 100% | Informational; only one no-match prediction |
| Adversarial no-match recall | 25% | Weakness; three of four non-skill requests over-routed |
| Observed top-5 recall | Insufficient (0 trustworthy rows) | Requires ≥30 rows |
| Warm route p50 / p95 | 2.1 ms / 3.8 ms | p95 <50 ms — pass |
| Cold process p50 / p95 | 19.8 ms / 54.2 ms | Informational |
| Full index time | 1.03 s | Informational |
| Registry size | 258,048 bytes | Informational |
| Real-catalog destructive operations | 0 | Must be 0 — pass |

Observed recall is not estimated. Existing audit trails record skill invocations but do not preserve a privacy-safe user request joined to an unambiguous expected skill, so manufacturing 30 labels would be misleading.

## Context accounting

| Context surface | Bytes | Versus 32,071-byte metadata baseline |
|---|---:|---:|
| Full discovery descriptions | 32,071 | Baseline |
| Router description only | 167 | 99.48% reduction |
| Entire router `SKILL.md` | 1,197 | 96.27% smaller |
| Router + selected top 3 bodies, p50 | 20,166 | 37.1% reduction |
| Router + selected top 3 bodies, p95 | 49,567 | 54.6% increase |
| Router + selected top 5 bodies, p50 | 27,967 | 12.8% reduction |
| Router + selected top 5 bodies, p95 | 79,181 | 146.9% increase |

The ≥80% gate is specifically the **modeled discovery-metadata reduction**, which passes at 99.48%. Total request context has a long tail. A three-candidate adapter is the smallest safe compromise because it retains 96.7% deterministic and 100% adversarial labeled recall while materially lowering median loaded-body bytes.

## Host-specific activation verdicts

| Host | Verdict | Evidence and limitation |
|---|---|---|
| Codex | **Context-saving when the managed catalog is activated; otherwise routing-only** | A dry-run against `~/.agents/skills` returned `mode=managed`, 87 symlink entries, and zero blockers. Codex still exposes skills from any other configured roots. No real activation occurred. See the current [Codex skill guidance](https://github.com/openai/codex/blob/main/codex-rs/skills/src/assets/samples/skill-creator/SKILL.md). |
| Claude Code | **Context-saving only in a router-only plugin/profile; otherwise routing-only** | Claude preloads available skill metadata. The Intermesh plugin declares only its router, but other installed skill plugins remain visible. See Anthropic's [Agent Skills overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview). |
| Hermes Agent | **Context-saving with a dedicated native router-only profile; otherwise routing-only** | Hermes indexes installed and external skill directories. Its native profiles/opt-out controls should isolate the router; adding canonical source roots as external directories would erase savings. See [Hermes Skills](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills). |

No host is marked unsupported, but none receives an automatic activation. Profile isolation remains an explicit, reversible operator action.

## Gate summary and follow-ups

All mandatory V0 gates pass except observed recall, which is correctly classified as insufficient rather than passed or failed. The production activation hold remains until Interspect accumulates at least 30 privacy-safe route outcomes.

Highest-value follow-ups:

1. Add an abstention threshold or calibrated no-skill classifier; weather, arithmetic, and joke requests currently receive lexical false positives.
2. Join `intermesh.route.v1` receipts to real outcomes in Interspect and rerun observed recall at 30+ trustworthy labels.
3. Measure actual host prompt tokens after an opt-in managed-profile trial; byte modeling is not a substitute for host telemetry.
4. Consider a bounded body-byte budget or two-stage selection if p95 selected-skill context remains above the baseline.

Reproduce with:

```bash
bash scripts/experiment.sh --check-gates
```
