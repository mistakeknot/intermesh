# intermesh

> See `AGENTS.md` for the complete development guide and ownership boundaries.

## Overview

Host-agnostic Go CLI for retrieval-gated Agent Skills discovery. Canonical `SKILL.md` and `intermesh.yaml` files are indexed into a rebuildable SQLite database. Thin host adapters call the CLI; they do not own routing logic.

## Quick Commands

```bash
go test ./...
go test -race ./...
go build -o /tmp/intermesh ./cmd/intermesh
/tmp/intermesh --help
```

## Design Decisions (Do Not Re-Ask)

- CLI-first; no MCP server.
- SQLite is a derived index, never the canonical source.
- Lexical/context routing ships before optional semantic reranking.
- Intermesh emits receipts; Interspect learns from outcomes.
- Intercore remains outside the write path.
- Host profile changes must support dry-run, snapshot, and restore.

