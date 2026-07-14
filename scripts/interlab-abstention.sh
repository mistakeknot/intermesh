#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/intermesh-abstention.XXXXXX")
trap 'rm -rf "$WORK"' EXIT

CASES=${INTERMESH_CASES:-$ROOT/testdata/routes/abstention-dev.jsonl}
BENCH_HOME=${INTERMESH_BENCH_HOME:-$HOME}
INDEX_SCRIPT=${INTERMESH_INDEX_SCRIPT:-$ROOT/scripts/index-codex-catalog.sh}

usage() {
    echo "usage: scripts/interlab-abstention.sh [--cases path] [--home path]" >&2
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --cases)
            [[ $# -ge 2 ]] || { usage; exit 2; }
            CASES=$2
            shift 2
            ;;
        --home)
            [[ $# -ge 2 ]] || { usage; exit 2; }
            BENCH_HOME=$2
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            usage
            exit 2
            ;;
    esac
done

[[ -f "$CASES" ]] || { echo "evaluation corpus not found: $CASES" >&2; exit 1; }
[[ -x "$INDEX_SCRIPT" ]] || { echo "index script is not executable: $INDEX_SCRIPT" >&2; exit 1; }

export GOCACHE=${GOCACHE:-$WORK/go-cache}
if [[ -z "${INTERMESH_BIN:-}" ]]; then
    INTERMESH_BIN="$WORK/intermesh"
    go build -o "$INTERMESH_BIN" "$ROOT/cmd/intermesh"
fi
if [[ -z "${INTERMESH_EVAL_BIN:-}" ]]; then
    INTERMESH_EVAL_BIN="$WORK/intermesh-eval"
    go build -o "$INTERMESH_EVAL_BIN" "$ROOT/cmd/intermesh-eval"
fi

if [[ "${INTERMESH_SKIP_TESTS:-0}" != "1" ]]; then
    go test "$ROOT/internal/route" "$ROOT/internal/eval" >&2
fi

database="$WORK/registry.db"
INTERMESH_BIN="$INTERMESH_BIN" "$INDEX_SCRIPT" --home "$BENCH_HOME" --db "$database" >/dev/null
"$INTERMESH_EVAL_BIN" \
    --db "$database" \
    --cases "$CASES" \
    --host codex \
    --limit 5 \
    --warmups 1 \
    --runs 5 > "$WORK/eval.json"

python3 - "$WORK/eval.json" <<'PY'
import json
import sys

report = json.load(open(sys.argv[1]))
metrics = report["metrics"]
values = {
    "no_match_recall": metrics["no_match_recall"],
    "no_match_precision": metrics["no_match_precision"],
    "top3_recall": metrics["top_3_recall"],
    "top5_recall": metrics["top_5_recall"],
    "mrr": metrics["mrr"],
    "warm_p95_micros": report["warm_latency"]["p95_micros"],
}
for name, value in values.items():
    print(f"METRIC {name}={value:.10g}")

failures = []
if values["top3_recall"] < 0.95:
    failures.append("top3_recall below 0.95")
if values["top5_recall"] < 0.95:
    failures.append("top5_recall below 0.95")
if values["no_match_precision"] < 0.80:
    failures.append("no_match_precision below 0.80")
if values["warm_p95_micros"] >= 50_000:
    failures.append("warm_p95_micros is not below 50000")
if failures:
    print("benchmark constraints failed: " + "; ".join(failures), file=sys.stderr)
    raise SystemExit(1)
PY
