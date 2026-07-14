# intermesh — Development Guide

## Canonical References

1. [`PHILOSOPHY.md`](./PHILOSOPHY.md) — direction for ideation and planning decisions.
2. `CLAUDE.md` — implementation details, architecture, testing, and release workflow.
3. [`docs/contracts/registry-v1.md`](./docs/contracts/registry-v1.md) — registry and relationship contract once implemented.

## Philosophy Alignment Protocol

Review [`PHILOSOPHY.md`](./PHILOSOPHY.md) during intake, planning, execution kickoff, review, and handoff. Planning outputs must include short `Alignment` and `Conflict/Risk` statements.

## Quick Reference

| Item | Value |
|---|---|
| Repository | `github.com/mistakeknot/intermesh` |
| CLI | `intermesh` |
| Runtime | Go |
| Persistence | SQLite, derived and rebuildable |
| Work tracking | Sylveste bead `sylveste-qn2c` |
| License | MIT |

## Boundaries

- Intermesh owns discovery, indexing, candidate generation, relationship resolution, and route receipts.
- Interskill owns skill creation and audit workflows.
- Interspect owns learned ranking policy and outcome calibration.
- Intercore owns orchestration state; Intermesh must not write Intercore state.
- Host adapters may invoke the CLI and load returned files but must not duplicate routing logic.

## Testing

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/intermesh
bash scripts/experiment.sh --check-gates
```

## Current Constraints

- The project name was previously reserved in Sylveste for a possible coordination substrate. The current explicit project decision consumes it for skill routing; update the old reservation when landing.
- A host cannot save context merely by receiving routing hints. Unselected skills must remain outside that host's automatic metadata catalog or be disabled through a reversible profile.
- Indexing untrusted skills must never execute their scripts or import their code.
- The experiment currently holds automatic catalog activation: observed recall has no trustworthy corpus yet, and no-match recall is 25%.
