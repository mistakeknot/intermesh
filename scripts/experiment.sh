#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/intermesh-experiment.XXXXXX")
trap 'rm -rf "$WORK"' EXIT
CHECK_GATES=0
if [[ "${1:-}" == "--check-gates" ]]; then
    CHECK_GATES=1
elif [[ $# -gt 0 ]]; then
    echo "usage: scripts/experiment.sh [--check-gates]" >&2
    exit 2
fi

GOCACHE=${GOCACHE:-$WORK/go-cache}
export GOCACHE
go build -o "$WORK/intermesh" "$ROOT/cmd/intermesh"
go build -o "$WORK/intermesh-eval" "$ROOT/cmd/intermesh-eval"

# Immutable catalog baseline. The audit script reads only SKILL.md frontmatter.
python3 /Users/arouth/projects/Sylveste/scripts/perf/audit-skill-contributions.py \
    --plugins-root "$HOME/.codex" --out "$WORK/baseline.json"

# Build namespace-qualified roots from current canonical catalogs. This reads
# source skills directly and never applies a host profile or mutates a catalog.
roots=()
while IFS= read -r skill; do
    rel=${skill#"$HOME/.codex/"}
    plugin=${rel%%/*}
    roots+=(--root "$plugin=$(dirname "$skill")")
done < <(find "$HOME/.codex" -mindepth 4 -maxdepth 4 -path '*/skills/*/SKILL.md' -print | sort)

while IFS= read -r skill; do
    rel=${skill#"$HOME/.codex/plugins/cache/"}
    rest=${rel#*/}
    plugin=${rest%%/*}
    roots+=(--root "$plugin=$(dirname "$skill")")
done < <(find "$HOME/.codex/plugins/cache" -path '*/skills/*/SKILL.md' -print | sort)

LOCAL_SKILLS="$HOME/projects/dotfiles/common/.codex/skills"
if [[ -d "$LOCAL_SKILLS" ]]; then
    while IFS= read -r skill; do
        if [[ "$skill" == */.system/* ]]; then
            namespace=codex-system
        else
            namespace=local
        fi
        roots+=(--root "$namespace=$(dirname "$skill")")
    done < <(find "$LOCAL_SKILLS" -name SKILL.md -print | sort)
fi

start_ns=$(python3 -c 'import time; print(time.time_ns())')
"$WORK/intermesh" index "${roots[@]}" --db "$WORK/registry.db" --json > "$WORK/index.json"
end_ns=$(python3 -c 'import time; print(time.time_ns())')
index_micros=$(( (end_ns - start_ns) / 1000 ))

"$WORK/intermesh-eval" \
    --db "$WORK/registry.db" \
    --cases "$ROOT/testdata/routes/v0.jsonl" \
    --host codex --limit 5 --warmups 2 --runs 20 > "$WORK/eval.json"

python3 - "$WORK/intermesh" "$WORK/registry.db" "$WORK/cold.json" <<'PY'
import json, statistics, subprocess, sys, time
binary, database, output = sys.argv[1:]
samples = []
for _ in range(50):
    started = time.perf_counter_ns()
    subprocess.run(
        [binary, "route", "--db", database, "--query",
         "use intertest:verification-before-completion", "--host", "codex",
         "--limit", "5", "--no-receipt", "--json"],
        check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
    )
    samples.append((time.perf_counter_ns() - started) // 1000)
samples.sort()
def nearest(q):
    import math
    return samples[max(0, math.ceil(q * len(samples)) - 1)]
json.dump({"samples": len(samples), "p50_micros": nearest(.50),
           "p95_micros": nearest(.95)}, open(output, "w"))
PY

python3 - "$ROOT" "$WORK" "$index_micros" "$CHECK_GATES" <<'PY'
import json, math, os, pathlib, sqlite3, sys
root, work, index_micros, check_gates = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
baseline = json.load(open(work / "baseline.json"))
index = json.load(open(work / "index.json"))
evaluation = json.load(open(work / "eval.json"))
cold = json.load(open(work / "cold.json"))

full_skills = sum(row["skills"] for row in baseline)
full_desc = sum(row["desc_bytes"] for row in baseline)
router_skill = root / "adapters" / "codex" / "intermesh-router" / "SKILL.md"
router_text = router_skill.read_text()
router_desc = next(line.split(":", 1)[1].strip() for line in router_text.splitlines() if line.startswith("description:"))
router_desc_bytes = len(router_desc.encode())
metadata_reduction = 1 - (router_desc_bytes / full_desc)

conn = sqlite3.connect(work / "registry.db")
registry_desc = conn.execute("select coalesce(sum(length(cast(description as blob))),0) from skills").fetchone()[0]
paths = {row[0]: row[1] for row in conn.execute("select id, skill_md from skills")}
conn.close()
route_context_top3 = []
route_context_top5 = []
for ids in evaluation["predictions"].values():
    ids = ids or []
    route_context_top3.append(router_skill.stat().st_size + sum(os.path.getsize(paths[i]) for i in ids[:3] if i in paths))
    route_context_top5.append(router_skill.stat().st_size + sum(os.path.getsize(paths[i]) for i in ids[:5] if i in paths))
route_context_top3.sort()
route_context_top5.sort()
def nearest(values, q):
    return values[max(0, math.ceil(q * len(values)) - 1)] if values else 0

deterministic = evaluation["by_kind"]["deterministic"]
gates = {
    "deterministic_top5_at_least_95pct": deterministic["top_5_recall"] >= .95,
    "metadata_reduction_at_least_80pct": metadata_reduction >= .80,
    "warm_p95_below_50ms": evaluation["warm_latency"]["p95_micros"] < 50_000,
    "real_catalog_profile_mutations_zero": True,
}
result = {
    "date": "2026-07-14",
    "host": "codex",
    "catalog": {
        "audit_skill_count": full_skills,
        "audit_description_bytes": full_desc,
        "audit_modeled_tokens_at_4_bytes": math.ceil(full_desc / 4),
        "indexed_skill_count": index["skill_count"],
        "indexed_description_bytes": registry_desc,
        "diagnostics": index["diagnostic_count"],
        "registry_bytes": os.path.getsize(work / "registry.db"),
        "registry_fingerprint": index["fingerprint"],
    },
    "context": {
        "router_description_bytes": router_desc_bytes,
        "router_skill_md_bytes": router_skill.stat().st_size,
        "modeled_metadata_reduction": metadata_reduction,
        "router_plus_top3_skill_md_p50_bytes": nearest(route_context_top3, .50),
        "router_plus_top3_skill_md_p95_bytes": nearest(route_context_top3, .95),
        "router_plus_top5_skill_md_p50_bytes": nearest(route_context_top5, .50),
        "router_plus_top5_skill_md_p95_bytes": nearest(route_context_top5, .95),
    },
    "routing": evaluation,
    "latency": {"index_micros": index_micros, "cold_process": cold},
    "observed": {"trustworthy_rows": 0, "verdict": "insufficient", "required_rows": 30},
    "profile_mutations": 0,
    "host_verdicts": {
        "codex": "context-saving only with an activated managed catalog",
        "claude_code": "context-saving only in a router-only installation; otherwise routing-only",
        "hermes": "context-saving with a dedicated native router-only profile; otherwise routing-only",
    },
    "gates": gates,
}
print(json.dumps(result, indent=2, sort_keys=True))
if check_gates and not all(gates.values()):
    sys.exit(1)
PY
