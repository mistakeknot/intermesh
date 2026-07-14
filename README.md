# intermesh

Retrieval-gated skill discovery for Codex, Claude Code, Hermes Agent, and other Agent Skills hosts.

## What this does

Intermesh scans canonical `SKILL.md` files and optional relationship manifests into a rebuildable SQLite registry. Its CLI returns a compact, ranked set of skills for a request, resolves declared requirements and conflicts, and emits routing receipts for downstream calibration.

The agent host remains responsible for loading and following selected instructions. Intermesh does not execute skill code, own orchestration state, or replace authoring and outcome-calibration systems.

## Status

V0 validated, with catalog activation intentionally opt-in. The live-catalog experiment measured 100% deterministic top-5 recall, 96.7% top-3 recall, 99.48% modeled always-on metadata reduction, and 3.2 ms warm p95 routing. Observed recall is still insufficient and no-match abstention needs improvement, so Intermesh does not automatically replace a host catalog. See the [experiment report](docs/reports/intermesh-v0-experiment.md).

## Build

```bash
go test ./...
go build -o ./bin/intermesh ./cmd/intermesh
```

The registry is derived local state. Index one or more canonical roots with an explicit namespace:

```bash
./bin/intermesh index \
  --root clavain=/path/to/clavain/skills \
  --root intertest=/path/to/intertest/skills
```

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

## Host adapters

- `adapters/codex/` — context-saving only after an explicit managed-catalog activation.
- `adapters/claude-code/` — context-saving only in a router-only plugin/profile.
- `adapters/hermes/` — context-saving through a dedicated native router-only profile.

All adapters request three candidates by default, then load the returned `SKILL.md` files in dependency order. The source catalog remains outside the host's automatic discovery roots.

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
