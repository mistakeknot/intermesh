# Intermesh Philosophy

## Purpose

Intermesh is a host-agnostic skill registry and routing CLI. It indexes canonical `SKILL.md` files and declarative relationships into a rebuildable local database, returns a small candidate set for a request, and leaves execution to the consuming agent host.

## North Star

Relevant-skill recall per metadata token exposed to the model.

## Working Priorities

1. Preserve routing recall while sharply reducing always-loaded metadata.
2. Keep the registry derived, local, inspectable, and safe to rebuild.
3. Integrate through narrow CLI and file contracts rather than host-specific runtime coupling.

## Brainstorming Doctrine

1. Start from observable routing failures and measured prompt cost.
2. Compare conservative, balanced, and aggressive options before adding machinery.
3. Prefer reversible experiments with explicit kill criteria.
4. Record assumptions, unknowns, and evidence gaps.

## Planning Doctrine

1. Define routing and context-budget acceptance criteria before implementation.
2. Separate mechanism from host policy and outcome calibration.
3. Build in testable, reversible slices with deterministic fallbacks.
4. Require a runtime integration path and durable receipt for every capability.

## Decision Filters

- Does this reduce metadata context without hiding relevant skills?
- Can the index be rebuilt entirely from canonical files?
- Does the CLI work without Codex, Claude Code, Hermes, or Intercore?
- Is learned policy owned by Interspect rather than smuggled into the registry?

## Evidence Base

- Sylveste skill listing audit, 2026-04-21: 156 skills and 32,330 description bytes.
- Sylveste search-ranking review, 2026-05-04: skill routing lacked candidate generation.
- Prefix-router experiment `sylveste-a4oj.9.1`: hint injection could not remove host-loaded metadata.
- Interspect skill-calibration plan, 2026-06-16: outcome learning belongs in Interspect.

