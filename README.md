# intermesh

Retrieval-gated skill discovery for Codex, Claude Code, Hermes Agent, and other Agent Skills hosts.

## What this does

Intermesh scans canonical `SKILL.md` files and optional relationship manifests into a rebuildable SQLite registry. Its CLI returns a compact, ranked set of skills for a request, resolves declared requirements and conflicts, and emits routing receipts for downstream calibration.

The agent host remains responsible for loading and following selected instructions. Intermesh does not execute skill code, own orchestration state, or replace authoring and outcome-calibration systems.

## Status

V0 validated, with catalog activation intentionally opt-in. The live-catalog experiment measured 100% deterministic top-5 recall, 96.7% top-3 recall, and 99.48% modeled always-on metadata reduction. A subsequent guarded Interlab campaign raised no-match recall from 25% to 100% on the original four-case set, from 20% to 73.3% on a 30-case development set, and from 5% to 40% on a locked holdout without reducing positive recall. A real Codex prompt-renderer smoke test of the isolated canary measured 98.65% less removable skill metadata, but observed route outcomes are still insufficient and indirect-paraphrase recall remains weak, so Intermesh does not automatically replace a host catalog. See the [V0 experiment](docs/reports/intermesh-v0-experiment.md), [campaign learnings](campaigns/intermesh-abstention-v1/learnings.md), and [Codex canary guide](docs/guides/codex-router-canary.md).

## Build

```bash
go test ./...
go build -o ./bin/intermesh ./cmd/intermesh
```

Install the current checkout and build a persistent registry from the active
Codex catalogs:

```bash
mkdir -p "$HOME/.local/bin"
GOBIN="$HOME/.local/bin" go install ./cmd/intermesh
export PATH="$HOME/.local/bin:$PATH"
scripts/index-codex-catalog.sh
intermesh doctor
```

The catalog helper discovers direct Codex plugins, versioned plugin-cache
skills, local dotfile skills, and Codex system skills. It only reads canonical
files and transactionally rebuilds the derived registry.

Set up a reversible Codex router-only canary without changing the normal
profile:

```bash
scripts/codex-canary.sh setup
scripts/codex-canary.sh compare --workspace "$PWD"
scripts/codex-canary.sh doctor --workspace "$PWD"
scripts/codex-canary.sh login
scripts/codex-canary.sh run --workspace "$PWD"
```

The canary isolates both `HOME` and `CODEX_HOME`, keeps authentication in its
own file-backed store, restores the real home for task subprocesses, and pins
router calls to its derived registry. See [Codex router-only canary](docs/guides/codex-router-canary.md)
before collecting or labeling sessions.

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

- `adapters/codex/` — context-saving in the isolated router-only canary; the normal profile remains untouched.
- `adapters/claude-code/` — context-saving only in a router-only plugin/profile.
- `adapters/hermes/` — context-saving through a dedicated native router-only profile.

All adapters request three candidates by default. Route JSON includes each
candidate's canonical frontmatter description, so the host can apply trigger
boundaries before loading complete bodies. Adapters fully load only applicable
skills and the complete `required_by` closure, filter potential conflicts
against that final selected set, and preserve dependency order. The source
catalog remains outside the host's automatic discovery roots.

## Testing and optimization

Run the fixed V0 experiment with `scripts/experiment.sh --check-gates`. For a
larger abstention-development corpus and Interlab-compatible `METRIC` output,
run `scripts/interlab-abstention.sh`. The benchmark fails closed if positive
top-3/top-5 recall, no-match precision, or warm latency cross their guardrails.

Run `bash tests/shell/test_codex_canary.sh` to verify isolated setup,
prompt-input analysis, session capture, the 30-label activation gate, and
lossless rollback behavior.

See [Testing and hill climbing](docs/guides/testing-and-hill-climbing.md) for
Interlab registration, campaign scope, baseline results, and holdout policy.

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
