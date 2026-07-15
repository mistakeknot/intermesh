# Codex Router-Only Canary — 2026-07-15

## Outcome

The isolated, router-only Codex profile passed its activation evidence gates on a
version-coherent 30-session cohort. This result does **not** activate or mutate
the normal Codex profile.

| Measure | Result | Gate |
|---|---:|---:|
| Trustworthy labeled sessions | 30 | 30 |
| Task score | 100% | ≥85% |
| Routing score | 100% | ≥90% |
| Route-receipt coverage | 100% | ≥90% |
| Removable non-system metadata reduction | 98.65% | ≥80% |
| Unexpected non-system skills in canary | 0 | 0 |

The cohort contained 26 skill-routing requests across distinct engineering
domains and four unrelated no-skill controls. Every request produced a private,
hashed route receipt. The no-skill controls returned empty candidate sets and
the agent abstained from loading a skill.

## Hill-climb history

An earlier 30-session exploratory campaign intentionally mixed router versions
while failures were being discovered. It finished with 100% task success but a
61.7% routing score, correctly producing a `hold` verdict. Those outcomes were
preserved rather than relabeled.

Observed failures drove public changes including progressive applicability
selection, substantive-query extraction, requirement and conflict parity across
adapters, lexical precision fixes, and discovery of both
`plugin/skills/SKILL.md` and `plugin/skills/<name>/SKILL.md` layouts. The last
fix restored `intermap:intermap` to the registry and placed it in the top three
for the real project-map prompt that exposed the omission.

## Context measurement

Codex prompt capture measured non-system skill metadata shrinking from 20,597
bytes across 118 exposed entries to 279 bytes for the Intermesh router. The
router then exposes bounded candidate metadata and loads only bodies selected
as applicable by the host agent.

The normal profile remained unchanged throughout. The activation cohort used a
separate canary root from the exploratory campaign so results represented one
registry and router generation without discarding historical evidence.

## Limitations and decision

The passing cohort deliberately covers many clear skill intents plus unrelated
abstention controls. It is not evidence that every indirect paraphrase routes
correctly. Earlier misses around planning, debugging versus TDD, and agent
paraphrasing remain useful regression targets for future hill-climbs.

Verdict: the router-only mechanism meets the defined canary gates and is safe to
offer as an explicit, reversible opt-in. Automatic activation remains off, and
the normal profile should not be changed without a separate user decision.
