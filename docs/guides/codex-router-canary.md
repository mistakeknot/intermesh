# Codex router-only canary

## Purpose

This canary measures whether Intermesh can replace Codex's removable automatic
skill metadata with one router while preserving useful task outcomes. It is a
separate launcher and state tree, not a mutation of the normal Codex profile.
Stopping the launcher immediately returns to normal Codex behavior.

Codex always contributes its bundled system skills. Repository-local
`.agents/skills` entries can also appear for a selected workspace. `doctor`
fails if any non-system skill other than the Intermesh router is visible, so a
workspace with extra repo skills is reported rather than silently counted as
router-only.

## Isolation model

`setup` creates `${INTERMESH_CANARY_ROOT:-~/.local/share/intermesh/codex-canary}`
with:

- an isolated `HOME` containing only the router under `.agents/skills`
- an isolated `CODEX_HOME` with file-backed Codex and MCP credentials
- a derived registry rebuilt from the normal canonical skill sources
- a private `intermesh` wrapper that pins routing to that registry
- append-only route, session, label, and context evidence

Codex skill and plugin discovery sees the isolated homes. Task subprocesses get
the normal home and path through `shell_environment_policy`, so Git and other
developer tools retain their ordinary user configuration. No auth file, plugin,
skill body, or normal Codex configuration is copied or changed. The workflow
uses no Tailscale or personal infrastructure.

## Set up and verify

Install Intermesh first, then create the canary:

```bash
mkdir -p "$HOME/.local/bin"
GOBIN="$HOME/.local/bin" go install ./cmd/intermesh
export PATH="$HOME/.local/bin:$PATH"

scripts/codex-canary.sh setup
scripts/codex-canary.sh compare --workspace /path/to/project
scripts/codex-canary.sh doctor --workspace /path/to/project
```

`compare` uses `codex debug prompt-input` for both profiles and saves the exact
model-visible skill-list comparison under the canary's `results/context.json`.
`doctor` checks the router link, isolated credential policy, registry health,
router presence, bundled-system count, and unexpected non-system skills.

Authentication is deliberately separate. Sign in once without copying the
normal profile's token:

```bash
scripts/codex-canary.sh login
```

Then start an interactive canary session:

```bash
scripts/codex-canary.sh run --workspace /path/to/project
```

## Collect 30 trustworthy sessions

For measured non-interactive work, give every real task a stable session ID and
optionally record the expected skill IDs:

```bash
scripts/codex-canary.sh exec \
  --session c01 \
  --workspace /path/to/project \
  --expected intertest:verification-before-completion \
  -- "review this change and verify it before reporting completion"
```

The command runs `codex exec --ephemeral --json`, retains the event stream,
joins any new Intermesh route receipts, and records Codex's actual
`turn.completed.usage`. The summary stores a normalized prompt hash, not the
prompt text. The raw local event file can still contain task output and command
details; treat the canary directory as private working data.

After inspecting the result, add a human label:

```bash
scripts/codex-canary.sh label \
  --session c01 \
  --task pass \
  --routing correct \
  --note "selected verification and completed the requested checks"
```

Task labels are `pass`, `partial`, or `fail`. Routing labels are `correct`,
`partial`, `wrong`, or `abstain`; use `abstain` only when no skill was the
correct routing decision. Labels are append-only, and the latest label for a
session is used in reports.

```bash
scripts/codex-canary.sh report
scripts/codex-canary.sh report --check-gates
```

Twenty sessions are enough for an interim report. The stricter existing policy
requires 30 fully labeled sessions for a `go` verdict. `--check-gates` exits
nonzero until all of these are true:

- campaign size is between 20 and 30, with at least 30 trustworthy labels
- every session is labeled
- weighted task score is at least 0.85
- weighted routing score is at least 0.90
- at least 90% of sessions emit a route receipt
- removable model-visible skill metadata falls by at least 80%
- no unexpected non-system skill is visible in the canary

Passing these gates authorizes a go/no-go recommendation; it does not mutate or
activate the normal profile automatically. Routing hill climbs must continue to
use locked positive-recall, precision, and latency gates from
[Testing and hill climbing](testing-and-hill-climbing.md).

## First real prompt-renderer measurement

On 2026-07-14, a disposable canary against Codex CLI 0.144.3 and the 170-skill
Intermesh registry produced:

| Model-visible skill list | Normal | Canary |
|---|---:|---:|
| Total skill entries | 123 | 6 |
| Bundled system entries | 5 | 5 |
| Removable/non-system entries | 118 | 1 router |
| Removable metadata bytes | 20,597 | 279 |
| Compact serialized prompt-input bytes | 37,504 | 12,871 |

The measured removable-metadata reduction was **98.65%** with zero unexpected
non-system skills. This proves host isolation and context reduction, not routing
quality. With zero observed canary sessions at measurement time, the activation
verdict remained `hold`.

## Roll back

The normal profile never changes, so the immediate rollback is simply to stop
using the canary launcher. To remove only the marked canary tree and its local
credentials/evidence:

```bash
scripts/codex-canary.sh uninstall --yes
```

`uninstall` refuses unmarked or unsafe roots. It does not delete canonical
skills, the normal Codex home, the normal user home, or project files.
