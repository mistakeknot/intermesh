# intermesh

Retrieval-gated skill discovery for Codex, Claude Code, Hermes Agent, and other Agent Skills hosts.

## What this does

Intermesh scans canonical `SKILL.md` files and optional relationship manifests into a rebuildable SQLite registry. Its CLI returns a compact, ranked set of skills for a request, resolves declared requirements and conflicts, and emits routing receipts for downstream calibration.

The agent host remains responsible for loading and following selected instructions. Intermesh does not execute skill code, own orchestration state, or replace authoring and outcome-calibration systems.

## Status

Experimental. The first milestone must prove routing recall, warm latency, and metadata-context reduction before the CLI is promoted for routine use.

## CLI

```bash
intermesh index --root intertest=/path/to/intertest/skills
intermesh route --query "review this pull request" --host codex --limit 5 --json
intermesh search "database migration"
intermesh resolve intertest:systematic-debugging
intermesh graph intertest:systematic-debugging
intermesh manifest validate /path/to/intermesh.yaml
intermesh profile plan --host codex --catalog ~/.agents/skills --router ./adapters/codex/intermesh-router --out plan.json
intermesh doctor
```

Profile activation and restore are explicit mutations and require `--yes`. V0 only manages removable catalog entries that are symlinks; regular files or directories force a routing-only plan.

Interskill optionally validates relationship manifests when this CLI is available. Interspect can ingest `intermesh.route.v1` receipts as decision evidence without fabricating outcome signals.

## Architecture

```text
SKILL.md + intermesh.yaml
          |
          v
  rebuildable SQLite index
          |
          v
 route/search/resolve/doctor CLI
          |
          +-- Codex router skill
          +-- Claude Code router skill
          +-- Hermes router skill
          +-- Interspect route receipts
```
