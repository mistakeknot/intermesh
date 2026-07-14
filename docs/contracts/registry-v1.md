# Intermesh Registry Contract V1

## Status

Version 1 is the initial local contract. Canonical information lives in skill packages; the SQLite registry is a disposable projection.

## Canonical skill package

Every package contains `SKILL.md` with YAML frontmatter delimited by `---` lines. `description` is required. When `name` is omitted, Intermesh applies the host-compatible default of the skill directory name; title-case display names are preserved while their registry ID segment is lowercased. Intermesh reads the file as data and never executes scripts, imports code, or expands template expressions while indexing.

An optional `intermesh.yaml` beside `SKILL.md` declares relationships and deterministic routing hints:

```yaml
version: 1
id: intertest:verification-before-completion
triggers:
  phrases:
    - verify before completion
  extensions:
    - .go
    - .py
  environments:
    - coding
requires:
  - intertest:test-driven-development
composes_with:
  - clavain:ship
conflicts_with: []
supersedes: []
```

Rules:

- `version` must equal `1`.
- IDs contain lowercase ASCII letters, digits, `_`, `-`, `.`, and at most one namespace separator `:`.
- An explicit manifest `id` wins. Otherwise, the ID is `<root namespace>:<frontmatter name>` when the root has a namespace, or the frontmatter name alone.
- Relationship IDs use the same syntax.
- Relative and absolute filesystem paths are not valid relationship IDs.
- Unknown fields are ignored by V1 readers. Writers must not destructively rewrite a manifest containing unknown fields.
- `requires` affects the selected set. `composes_with`, `conflicts_with`, and `supersedes` are explanatory in V1.

## Root declaration

A root consists of a canonical absolute directory and an optional namespace. Discovery recursively locates files named exactly `SKILL.md`. Resolved files must remain beneath the declared root. Duplicate canonical paths are deduplicated; duplicate skill IDs are a registry-generation error.

## Diagnostics

Indexing emits structured diagnostics:

```json
{
  "path": "/abs/path/SKILL.md",
  "severity": "warning",
  "code": "frontmatter.missing_name",
  "message": "name is required"
}
```

Invalid individual packages are omitted. Errors that make the generation ambiguous, including duplicate IDs, abort replacement and preserve the prior generation.

## Registry fingerprint

The fingerprint is `sha256:<hex>` over sorted normalized canonical records. It includes IDs, names, descriptions, canonical paths, body hashes, triggers, and edges. It excludes timestamps, filesystem enumeration order, SQLite row IDs, and source mtimes.

## Route result V1

```json
{
  "version": 1,
  "query": "verify this change",
  "host": "codex",
  "registry_fingerprint": "sha256:...",
  "candidates": [
    {
      "id": "intertest:verification-before-completion",
      "skill_md": "/abs/path/SKILL.md",
      "score": 12.5,
      "reasons": ["phrase:verify", "lexical:completion"],
      "selected_by": "rank",
      "required_by": []
    }
  ],
  "warnings": []
}
```

The result is bounded by the requested limit before requirements are expanded. Required candidates use `selected_by: "requirement"` and name their parent IDs under `required_by`.

## Route receipt V1

Receipts are append-only JSON Lines records with `event: "intermesh.route.v1"`, a unique route ID, UTC timestamp, query hash, cwd, host, registry fingerprint, requested limit, candidate IDs and score components, warnings, and latency in microseconds. Raw query text is absent by default.

Intermesh records predictions only. Interspect or another evidence system attaches outcomes and publishes bounded calibration overlays through a separate versioned contract.

## Compatibility

Adding optional fields is compatible. Changing field meaning, required fields, ID rules, selection semantics, or receipt privacy defaults requires a new major contract version.
