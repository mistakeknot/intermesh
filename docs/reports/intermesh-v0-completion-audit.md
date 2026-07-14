---
artifact_type: completion-audit
bead: sylveste-qn2c
date: 2026-07-14
status: ready-to-land
---
# Intermesh V0 Completion Audit

This audit maps the persistent goal and implementation-plan must-haves to direct evidence. Publication, pushes, and bead closure remain landing operations rather than implementation gaps.

## Goal requirements

| Requirement | Evidence | Verdict |
|---|---|---|
| Host-agnostic, kernel-grade CLI | `cmd/intermesh`, documented exit codes, JSON/human output, Go-only core, no host SDK dependency | Met |
| Index `SKILL.md` plus versioned relationships | `internal/skill`, `internal/manifest`, `intermesh manifest validate`, hostile parser/manifest tests | Met |
| Rebuildable SQLite registry | Embedded schema, deterministic fingerprint, atomic `Replace`, duplicate-ID preservation tests | Met |
| Compact routing and search | `route`, `search`, limit enforcement, stable tie-breaks, score reasons and components | Met |
| Dependency resolution and graph inspection | `resolve`, `graph`, transitive ordering, diamond deduplication, exact cycle/missing-edge tests, conflict warnings | Met |
| Indexing and diagnostics | `index`, `doctor`, persisted structured diagnostics, live-catalog run with 170 skills and zero diagnostics | Met |
| Clean Interskill integration | Optional `command -v intermesh` authoring/audit behavior plus standalone contract tests; absence remains a no-op | Met |
| Preserve Intercore boundary | Intermesh writes only its derived registry/receipts; architecture and contracts explicitly exclude Intercore state | Met |
| Codex, Claude Code, Hermes adapters | Three router skills, plugin manifest, adapter descriptions under 300 characters, host-specific READMEs | Met |
| Routing-quality/context experiment | Reproducible `scripts/experiment.sh`, 40-case corpus, evaluation package/CLI, experiment report | Met |
| Production-worthy tests and documentation | Unit/race suites, cross-repo tests, CI workflow, mission/philosophy/conventions/contracts/reports | Met |
| Land the smallest safe implementation | Six logical Intermesh commits plus dependency and Sylveste commits; automatic profile activation intentionally excluded | Ready; remote/push pending |

## Plan truths

| Must-have truth | Direct proof |
|---|---|
| Multiple roots produce stable registry fingerprint without executing skills | Discovery reads files only; deterministic fingerprint test; repeated live run fingerprint `sha256:5465725...` |
| Bounded candidates with inspectable scores | Rank limit test; JSON `components` map sums to aggregate score |
| Requirements explicit; cycles/conflicts never silent | Graph unit tests and warnings/errors |
| Failed rebuild preserves prior generation | Duplicate-ID atomicity test |
| Registry is disposable derived state | Canonical paths/hashes/manifests are fingerprint inputs; experiment rebuilds from scratch |
| Profile changes dry-run, snapshotted, reversible | Plan/non-mutation, integrity hash, drift refusal, idempotence, crash-recovery, traversal, restore tests |
| Routes emit outcome-attachable receipts | `intermesh.route.v1`, private hashed query default, 0600 file test |
| Host savings are not generalized | Separate Codex/Claude Code/Hermes verdicts in adapter docs and experiment report |

## Experiment gates

| Gate | Result |
|---|---:|
| Deterministic top-5 recall ≥95% | 100% |
| Deterministic operational top-3 recall | 96.7% |
| Observed top-5 recall ≥85% at 30+ rows | Insufficient: 0 trustworthy rows; not passed or fabricated |
| Modeled metadata reduction ≥80% | 99.48% |
| Warm p95 <50 ms | Passed; 3.2 ms report run and 5.9 ms post-commit verification run |
| Real catalog destructive operations | 0 |

## Cross-repository evidence

- Interskill `ff963e6`: optional manifest authoring/validation, tests pass.
- Interspect `ecaafab`: route-receipt ingestion and decision/outcome separation; all shell tests pass.
- Interspect `8e01879`: macOS test-count portability fix discovered by the full verification run.
- Sylveste `86c0741f`: canonical name reservation and architecture inventory updated without staging unrelated user changes.

## Deliberate holds, not missing scope

- Automatic real-catalog activation is held because observed recall has no trustworthy corpus and adversarial no-match recall is 25%.
- The adapters request three skills because five-body loading has a 79 KB p95 context tail.
- Semantic reranking and learned priors remain future work; Interspect owns calibration.
- No installer mutates host catalogs. The experiment selected explicit operator-run profiles as the safe V0 wiring.

## Landing remainder

1. Create the `mistakeknot/intermesh` GitHub repository with user-selected visibility.
2. Push Intermesh, Interskill, Interspect, and Sylveste commits.
3. Run post-push status/CI checks.
4. Close `sylveste-qn2c` through the canonical gate and push Beads state.
