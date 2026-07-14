# Conventions

- Project slug and binary: `intermesh`.
- Canonical skill source: `SKILL.md` YAML frontmatter plus body.
- Optional relationship source: `intermesh.yaml` beside `SKILL.md`.
- Machine output: versioned JSON on stdout; diagnostics on stderr.
- Persistent index: `${XDG_DATA_HOME:-~/.local/share}/intermesh/registry.db`.
- Route receipts: `${XDG_STATE_HOME:-~/.local/state}/intermesh/routes.jsonl`.
- Tests use temporary directories and databases; never mutate real host catalogs.

