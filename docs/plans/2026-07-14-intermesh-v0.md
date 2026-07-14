---
artifact_type: plan
bead: sylveste-qn2c
stage: design
requirements:
  - F1: Canonical skill and relationship discovery
  - F2: Rebuildable SQLite registry
  - F3: Compact deterministic routing
  - F4: Dependency and conflict resolution
  - F5: Diagnostics and safe host profiles
  - F6: Codex, Claude Code, and Hermes adapters
  - F7: Interskill and Interspect integration contracts
  - F8: Routing-quality and context-reduction experiment
---
# Intermesh V0 Implementation Plan

> **For Codex/Claude:** Implement sequentially with test-driven development. Do not delegate tasks; the workspace AGENTS.md requires main-thread execution for this project.

**Bead:** `sylveste-qn2c`

**Goal:** Land a production-worthy, host-agnostic CLI that retrieves a compact set of relevant Agent Skills from an out-of-prompt catalog, safely resolves relationships, and proves the resulting context savings and routing quality.

**Architecture:** Canonical `SKILL.md` and adjacent `intermesh.yaml` files are parsed into a transactionally replaced SQLite index. A deterministic ranker combines exact identifiers, explicit triggers, environment/filetype filters, lexical relevance, and optional calibrated priors. The CLI emits versioned JSON plus append-only route receipts. Thin host adapters keep only the router skill automatically visible and load selected skill files through the host's normal filesystem tools.

**Tech Stack:** Go 1.24+, `modernc.org/sqlite`, `gopkg.in/yaml.v3`, standard-library CLI and JSON packages, shell smoke tests, Markdown Agent Skills.

**Prior Learnings:**

- `sylveste-a4oj.9.1` proved UserPromptSubmit hints cannot remove metadata already loaded by the host; profile isolation is part of the product, not optional polish.
- `docs/research/2026-04-21-skill-listing-audit.md` measured 156 skills and 32,330 description bytes; use its script and methodology for the baseline.
- `docs/research/flux-drive/2026-05-04-target/fd-search-engine-ranking.md` recommends cheap candidate generation before expensive model judgment.
- `docs/plans/2026-06-16-interspect-skill-calibration.md` establishes that learned skill outcomes belong in Interspect.
- `docs/solutions/patterns/hybrid-cli-plugin-architecture-20260223.md` requires the CLI to own data and logic while skills remain thin adapters.

**Alignment:** The plan improves relevant capability per context token through a small, inspectable Unix boundary and preserves strong policy/mechanism separation.

**Conflict/Risk:** Host-level catalog isolation is not equally supported across Codex, Claude Code, and Hermes. Each adapter must state and test its real guarantee rather than inheriting a generic context-savings claim.

---

## Must-Haves

**Truths**

- A user can index multiple skill roots and receive a stable registry fingerprint without executing any skill code.
- A request returns no more than the configured candidate limit, with score components and concise reasons inspectable in JSON.
- Required skills are added after ranking; cycles and conflicts are explicit errors or warnings, never silent behavior.
- A failed rebuild leaves the previous registry usable.
- The registry can be deleted and reconstructed without information loss because all canonical metadata lives beside the skills.
- Profile operations are dry-run by default, snapshot before mutation, and restore exactly what they changed.
- Every real route can emit a receipt suitable for Interspect outcome attachment.
- Experiment reports distinguish modeled context savings from savings actually achievable by each host.

**Artifacts**

- `cmd/intermesh/main.go` exposes the CLI.
- `internal/skill/` discovers and parses skill packages.
- `internal/registry/` owns schema and atomic rebuilds.
- `internal/route/` ranks and explains candidates.
- `internal/graph/` resolves declared relationships.
- `internal/profile/` plans, applies, and restores host catalogs.
- `adapters/{codex,claude-code,hermes}/` contains thin router skills and host notes.
- `docs/contracts/registry-v1.md` defines the canonical manifest, JSON, and receipt contracts.
- `docs/reports/intermesh-v0-experiment.md` records the measured result and activation verdict.

**Key Links**

- `intermesh index` parses canonical files and calls one atomic `registry.Replace` transaction.
- `intermesh route` queries the registry, calls `route.Rank`, then `graph.Resolve`, then optionally `receipt.Append`.
- Host adapters invoke `intermesh route --json` and read only returned `skill_md` paths.
- Interskill invokes `intermesh manifest validate` when the binary is available; absence remains a graceful no-op.
- Interspect consumes route receipts but remains the only owner of learned boosts and outcomes.

---

### Task 1: Freeze Registry, Manifest, Route, and Receipt Contracts

**Files:**

- Create: `docs/contracts/registry-v1.md`
- Create: `testdata/skills/minimal/SKILL.md`
- Create: `testdata/skills/depends/SKILL.md`
- Create: `testdata/skills/depends/intermesh.yaml`
- Create: `testdata/skills/conflict/SKILL.md`
- Create: `testdata/skills/conflict/intermesh.yaml`
- Create: `testdata/skills/invalid/SKILL.md`

**Step 1: Define `intermesh.yaml` version 1**

Use this shape and declare unknown fields forward-compatible:

```yaml
version: 1
id: intertest:verification-before-completion
triggers:
  phrases: ["verify before completion"]
  extensions: [".go", ".py"]
  environments: ["coding"]
requires: ["intertest:test-driven-development"]
composes_with: ["clavain:ship"]
conflicts_with: []
supersedes: []
```

The default ID is `<namespace>:<frontmatter.name>` when a namespace can be inferred from a root declaration; otherwise it is the frontmatter name.

**Step 2: Define versioned JSON envelopes**

`route --json` must emit:

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

`routes.jsonl` adds `route_id`, UTC timestamp, cwd, requested limit, returned IDs, latency, and score components. It contains no prompt body unless the user opts in.

**Step 3: Write hostile fixtures**

Cover malformed frontmatter, a path escaping its root, duplicate IDs, dependency cycles, missing requirements, and symmetric/asymmetric conflicts.

<verify>
- run: `test -f docs/contracts/registry-v1.md && find testdata/skills -name SKILL.md | grep -q .`
  expect: exit 0
</verify>

### Task 2: Discover and Parse Skills Without Executing Them

**Files:**

- Create: `go.mod`
- Create: `internal/skill/model.go`
- Create: `internal/skill/frontmatter.go`
- Create: `internal/skill/discover.go`
- Create: `internal/skill/frontmatter_test.go`
- Create: `internal/skill/discover_test.go`

**Step 1: Write failing table-driven parser tests**

Test valid quoted and folded YAML descriptions, optional `name`, UTF-8 BOM, CRLF, missing closing delimiter, description over 1024 characters, and body hashes. Parsing returns structured diagnostics rather than panicking.

**Step 2: Run the focused tests and confirm RED**

Run: `go test ./internal/skill -run 'Test(Parse|Discover)' -v`

Expected: compile failure because parser/discovery symbols do not exist.

**Step 3: Implement the minimal parser and discovery walker**

Required public shapes:

```go
type Skill struct {
    ID, Namespace, Name, Description string
    SkillMD, Directory, BodyHash     string
    Manifest                         Manifest
}

type Diagnostic struct {
    Path, Code, Message string
    Severity            string
}

func Parse(path, namespace string) (Skill, []Diagnostic)
func Discover(ctx context.Context, roots []Root) ([]Skill, []Diagnostic, error)
```

Walk only declared roots, do not execute or import files, reject resolved paths outside a root, and sort by canonical absolute path before returning.

**Step 4: Run parser/discovery tests**

Run: `go test ./internal/skill -v`

Expected: PASS.

<verify>
- run: `go test ./internal/skill -v`
  expect: exit 0
</verify>

### Task 3: Build a Transactionally Replaceable SQLite Registry

**Files:**

- Create: `internal/registry/schema.sql`
- Create: `internal/registry/store.go`
- Create: `internal/registry/index.go`
- Create: `internal/registry/store_test.go`
- Create: `internal/registry/index_test.go`

**Step 1: Write failing atomicity and determinism tests**

Tests must prove:

- Identical inputs produce the same `sha256:` fingerprint.
- A duplicate ID aborts replacement and preserves the previous generation.
- Invalid individual skills are reported and omitted without aborting valid entries.
- Removing a source skill removes it after the next successful rebuild.
- Readers observe either the old or new generation, never a partial generation.

**Step 2: Run tests and confirm RED**

Run: `go test ./internal/registry -v`

Expected: compile failure for missing store and indexer.

**Step 3: Implement schema and replacement transaction**

Tables: `registry_meta`, `skills`, `triggers`, `edges`, `roots`, and `diagnostics`. Enable WAL, foreign keys, and busy timeout. Store normalized descriptions, body hashes, source mtimes/sizes for diagnostics, and all declared relationship edges. Compute the fingerprint from sorted canonical records, not SQLite row order or timestamps.

**Step 4: Pass registry tests and race tests**

Run: `go test -race ./internal/registry -v`

Expected: PASS.

<verify>
- run: `go test -race ./internal/registry -v`
  expect: exit 0
</verify>

### Task 4: Rank Compact Candidates and Emit Receipts

**Files:**

- Create: `internal/route/tokenize.go`
- Create: `internal/route/score.go`
- Create: `internal/route/route.go`
- Create: `internal/route/route_test.go`
- Create: `internal/receipt/writer.go`
- Create: `internal/receipt/writer_test.go`
- Create: `testdata/routes/deterministic.jsonl`

**Step 1: Write failing routing tests**

Cover exact ID/name, namespaced command, phrase trigger, extension, environment, lexical overlap, negative environment filter, stable tie-breaking, limit enforcement, and no-match behavior. Each expected candidate must appear within top 5 rather than requiring rank 1.

**Step 2: Implement explainable baseline scoring**

Use fixed, documented components:

```text
exact ID/name       +100
namespaced command   +80
declared phrase      +40
extension match      +20
environment match    +10
lexical term          +IDF-weighted overlap
calibrated prior      bounded to [-5,+5]
```

Normalize Unicode and case, discard a checked-in stopword set, and use stable ID ordering for ties. Do not call an LLM or embedding service in V0.

**Step 3: Add privacy-minimal receipts**

Append with `O_APPEND`, create parent directories with `0700`, file with `0600`, and omit raw query text by default. Hash the normalized query; allow `--receipt-query=plain` only explicitly.

**Step 4: Pass route and receipt tests**

Run: `go test -race ./internal/route ./internal/receipt -v`

Expected: PASS.

<verify>
- run: `go test -race ./internal/route ./internal/receipt -v`
  expect: exit 0
</verify>

### Task 5: Resolve Requirements and Surface Graph Problems

**Files:**

- Create: `internal/manifest/manifest.go`
- Create: `internal/manifest/manifest_test.go`
- Create: `internal/graph/resolve.go`
- Create: `internal/graph/resolve_test.go`

**Step 1: Write failing relationship tests**

Test transitive requirements, diamond dependencies, deterministic topological ordering, missing requirement, cycle path reporting, direct and transitive conflict reporting, `composes_with`, and `supersedes` visibility.

**Step 2: Implement strict manifest validation**

Reject unsupported major versions and invalid identifiers. Preserve unknown fields for forward-compatible reads but never write them back destructively. `requires` changes selection; other edge types are explanatory in V0.

**Step 3: Implement resolver**

`Resolve(selected []Candidate)` returns ordered candidates plus warnings. A cycle returns an error containing the exact cycle. A missing required skill returns an error. Conflicts return warnings and keep both candidates unless a host policy explicitly chooses otherwise.

**Step 4: Pass relationship tests**

Run: `go test -race ./internal/manifest ./internal/graph -v`

Expected: PASS.

<verify>
- run: `go test -race ./internal/manifest ./internal/graph -v`
  expect: exit 0
</verify>

### Task 6: Wire the CLI and Diagnostics

**Files:**

- Create: `cmd/intermesh/main.go`
- Create: `internal/cli/app.go`
- Create: `internal/cli/index.go`
- Create: `internal/cli/route.go`
- Create: `internal/cli/search.go`
- Create: `internal/cli/resolve.go`
- Create: `internal/cli/graph.go`
- Create: `internal/cli/doctor.go`
- Create: `internal/cli/cli_test.go`

**Step 1: Write failing black-box CLI tests**

Use a temporary HOME/XDG data directory. Assert help, exit codes, compact JSON schema, human output, stale-index warning, unsupported schema error, invalid root diagnostics, and an index→route→resolve happy path.

**Step 2: Implement commands with standard-library `flag.FlagSet`**

Commands:

```text
intermesh index [--root namespace=path] [--db path] [--json]
intermesh route --query text [--host host] [--cwd path] [--limit 5] [--json]
intermesh search query [--limit 20] [--json]
intermesh resolve id... [--json]
intermesh graph id [--depth 2] [--json]
intermesh manifest validate path [--json]
intermesh doctor [--host host] [--json]
```

Structured output stays on stdout; diagnostics and progress stay on stderr. Exit codes: `0` success, `2` invalid invocation/input, `3` unhealthy registry, `4` unresolved graph, `5` unsafe profile operation.

**Step 3: Run CLI tests and build**

Run: `go test -race ./internal/cli -v && go build -o /tmp/intermesh ./cmd/intermesh`

Expected: PASS and binary created.

<verify>
- run: `go test -race ./internal/cli -v`
  expect: exit 0
- run: `go build -o /tmp/intermesh ./cmd/intermesh`
  expect: exit 0
- run: `/tmp/intermesh --help`
  expect: contains "route"
</verify>

### Task 7: Add Safe Profiles and Three Thin Host Adapters

**Files:**

- Create: `internal/profile/model.go`
- Create: `internal/profile/plan.go`
- Create: `internal/profile/apply.go`
- Create: `internal/profile/restore.go`
- Create: `internal/profile/profile_test.go`
- Create: `adapters/codex/intermesh-router/SKILL.md`
- Create: `adapters/codex/README.md`
- Create: `adapters/claude-code/intermesh-router/SKILL.md`
- Create: `adapters/claude-code/README.md`
- Create: `adapters/hermes/intermesh-router/SKILL.md`
- Create: `adapters/hermes/README.md`
- Create: `skills/router/SKILL.md`
- Create: `.claude-plugin/plugin.json`

**Step 1: Write failing profile safety tests**

Tests use fake host catalogs and prove dry-run default, explicit `--apply`, snapshot manifest with hashes, no traversal outside catalog roots, idempotent activation, refusal on drift, and exact restore. Simulate a crash after snapshot and verify recovery.

**Step 2: Implement host-neutral profile plans**

Add:

```text
intermesh profile plan --host <host> --catalog <path>
intermesh profile apply --plan <file> --yes
intermesh profile restore --snapshot <id> --yes
```

Profile mechanism creates a managed catalog containing the router plus configured always-on skills. It never deletes source skill repositories. If the host cannot point at the managed catalog without mutating unsupported state, the plan reports `routing_only` and does not claim context reduction.

**Step 3: Write adapters**

Each adapter tells the host to:

1. Call `intermesh route` with the current request, host, cwd, and JSON.
2. Read every returned `skill_md` fully, in order.
3. Obey host-native skill instructions normally.
4. Surface resolver warnings.
5. Record outcome through Interspect when available.

Keep each adapter description under 300 characters. Do not embed the catalog or graph in the adapter.

**Step 4: Verify adapters against current official host behavior**

Document exact catalog discovery and any limitations in each README. Claude Code and Codex must be verified from their current official docs/local behavior; Hermes must be verified against the current NousResearch source/docs.

<verify>
- run: `go test -race ./internal/profile -v`
  expect: exit 0
- run: `find adapters -name SKILL.md -print | wc -l`
  expect: contains "3"
- run: `python3 -c 'import json; json.load(open(".claude-plugin/plugin.json"))'`
  expect: exit 0
</verify>

### Task 8: Integrate Interskill Authoring and Interspect Receipts

**Files:**

- Create: `docs/contracts/interskill-integration.md`
- Create: `docs/contracts/interspect-receipts-v1.md`
- Modify in canonical Interskill repo: `skills/skill/SKILL.md`
- Modify in canonical Interskill repo: `skills/audit/SKILL.md`
- Add tests in canonical Interskill repo for graceful optional invocation
- Modify or add the smallest canonical Interspect ingestion path selected after inspecting its shipped state
- Add Interspect fixture/tests for `intermesh.route.v1`

**Step 1: Freeze optional integration behavior**

Interskill runs `command -v intermesh` first. If present, creation offers an adjacent manifest and runs `intermesh manifest validate`; audit reports registry/manifest problems. If absent, existing authoring behavior is unchanged.

**Step 2: Freeze receipt ownership**

Intermesh records the prediction: query hash, candidates, scores, latency, registry fingerprint. Interspect records actual outcomes and calibrated adjustments. Intermesh reads only an explicit, bounded calibration overlay exported by Interspect; it never learns directly in its registry database.

**Step 3: Implement and test the cross-repo changes**

Clone canonical repos into `/Users/arouth/projects/` if absent. Preserve their local AGENTS.md workflows, use the shared bead ID, and commit dependency changes before Intermesh integration changes.

<verify>
- run: `rg -n "intermesh manifest validate" /Users/arouth/projects/interskill/skills`
  expect: contains "intermesh manifest validate"
- run: `rg -n "intermesh.route.v1" /Users/arouth/projects/interspect`
  expect: contains "intermesh.route.v1"
</verify>

### Task 9: Run the Routing and Context Experiment

**Files:**

- Create: `scripts/experiment.sh`
- Create: `internal/eval/eval.go`
- Create: `internal/eval/eval_test.go`
- Create: `testdata/routes/observed.jsonl` from privacy-safe labeled receipts when available
- Create: `docs/reports/intermesh-v0-experiment.md`

**Step 1: Capture immutable baseline**

Run Sylveste's `scripts/perf/audit-skill-contributions.py` against the current catalog. Record unique skill count, description bytes, modeled tokens, host, and date. Do not overwrite the source audit.

**Step 2: Build evaluation sets**

- Deterministic: explicit namespaced commands, file extensions, and exact phrases with unambiguous expected skills.
- Observed: join historical user requests to actual skill invocations only where the audit trail supports the label; omit ambiguous/no-label rows.
- Adversarial: near-neighbor skills, conflicting descriptions, no-skill requests, and renamed/superseded skills.

**Step 3: Measure**

Report top-1 and top-5 recall, MRR, no-match precision, warm/cold p50/p95 latency, index time, registry size, full metadata bytes, router-only bytes, and router-plus-top-k bytes.

**Step 4: Apply gates**

- Deterministic top-5 recall ≥95%.
- Observed top-5 recall ≥85% when at least 30 trustworthy rows exist; otherwise label the result insufficient rather than passing.
- Modeled metadata reduction ≥80%.
- Warm p95 route latency <50 ms.
- Zero destructive profile operations against real catalogs during the experiment.

**Step 5: Write host-specific verdicts**

For each host, state `context-saving`, `routing-only`, or `unsupported`, with evidence. Do not generalize a Codex result to Claude Code or Hermes.

<verify>
- run: `go test -race ./internal/eval -v`
  expect: exit 0
- run: `bash scripts/experiment.sh --check-gates`
  expect: exit 0
- run: `rg -n "top-5|metadata reduction|p95|Codex|Claude Code|Hermes" docs/reports/intermesh-v0-experiment.md`
  expect: exit 0
</verify>

### Task 10: Production Verification and Landing

**Files:**

- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `CLAUDE.md`
- Modify: `docs/brainstorms/2026-07-14-retrieval-gated-skill-mesh.md`
- Modify in Sylveste: `docs/brainstorms/2026-04-29-bead-aware-cross-agent-coordination.md`
- Modify in Sylveste: architecture/module inventory and installer/profile wiring selected by the experiment verdict
- Create: `.github/workflows/ci.yml`

**Step 1: Run complete verification**

```bash
go test ./...
go test -race ./...
go vet ./...
go build -o /tmp/intermesh ./cmd/intermesh
/tmp/intermesh doctor --json
```

**Step 2: Perform completion audit**

Map every objective requirement and plan must-have to direct evidence: tests, CLI output, adapter files, cross-repo integration, experiment report, and host-specific catalog proof. Missing or indirect evidence keeps the bead and goal open.

**Step 3: Review**

Review security boundaries (untrusted skill paths, symlinks, YAML bombs, profile restore), correctness (atomic replacement and graph cycles), and product claims (modeled versus realized token savings).

**Step 4: Create/publish repository only with user-approved remote authority**

The local repository may be committed during development. Before `gh repo create`, ask whether the remote should be private or public, as required by the project-onboarding workflow. Add it to Sylveste only after the experiment passes.

**Step 5: Land dependency-first**

Commit/push Interspect and Interskill changes first, then Intermesh, then Sylveste inventory/installer changes. Close `sylveste-qn2c` only after all remotes are up to date and the real-host smoke checks pass.

<verify>
- run: `go test ./... && go test -race ./... && go vet ./...`
  expect: exit 0
- run: `git status --short`
  expect: exit 0
</verify>

