# Interspect Route Receipt Contract

Status: V1

Intermesh emits routing decisions. Interspect owns outcome attachment, longitudinal evidence, calibration, canaries, and any learned routing prior. A routing decision is not evidence that a selected skill succeeded.

## Input

Interspect consumes append-only JSONL records whose `event` is `intermesh.route.v1`, as defined in [registry-v1.md](registry-v1.md). The default source is:

```text
${XDG_STATE_HOME:-$HOME/.local/state}/intermesh/routes.jsonl
```

Raw request text is absent by default. The stable join key is `route_id`; `query_hash` supports grouping without disclosing the prompt.

## Normalization

Interspect expands each candidate into one `evidence` row:

| Evidence field | Receipt source |
|---|---|
| `source_kind` | `skill` |
| `event` | `skill_route` |
| `source` | `candidates[].id` |
| `session_id` | `route_id` until a host session join is available |
| `source_event_id` | deterministic `route_id`, rank, and skill ID tuple |
| `project` | `cwd` |
| `source_table` | `intermesh_routes` |
| `context` | route ID, host, query hash, registry fingerprint, rank, score, selection source, requirements, warnings, and latency |

Ingestion is idempotent and uses an Intermesh-specific watermark so newer route receipts cannot suppress older audit-log records.

## Outcome boundary

Receipt ingestion writes no `skill_signals` row. Selection, loading, invocation, and success are distinct events. Interspect may later attach an observed outcome using `route_id` plus host/session evidence; until then the route remains decision-only evidence and cannot change production routing weights.

Intermesh may eventually read a versioned, user-approved calibration export. It must never write Interspect's database, overlays, canaries, or routing overrides.
