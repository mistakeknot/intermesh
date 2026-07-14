#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/intermesh-routing-v2.XXXXXX")
trap 'rm -rf "$WORK"' EXIT

bash "$ROOT/scripts/interlab-abstention.sh" > "$WORK/development.metrics"
bash "$ROOT/scripts/interlab-abstention.sh" \
    --cases "$ROOT/testdata/routes/observed-regressions-dev.jsonl" \
    --report-only > "$WORK/observed.metrics"

cat "$WORK/development.metrics"
awk '
    $1 == "METRIC" && $2 ~ /^top3_recall=/ {
        sub(/^top3_recall=/, "", $2)
        print "METRIC observed_top3_recall=" $2
    }
    $1 == "METRIC" && $2 ~ /^top5_recall=/ {
        sub(/^top5_recall=/, "", $2)
        print "METRIC observed_top5_recall=" $2
    }
    $1 == "METRIC" && $2 ~ /^mrr=/ {
        sub(/^mrr=/, "", $2)
        print "METRIC observed_mrr=" $2
    }
    $1 == "METRIC" && $2 ~ /^no_match_precision=/ {
        sub(/^no_match_precision=/, "", $2)
        print "METRIC observed_no_match_precision=" $2
    }
    $1 == "METRIC" && $2 ~ /^no_match_recall=/ {
        sub(/^no_match_recall=/, "", $2)
        print "METRIC observed_no_match_recall=" $2
    }
' "$WORK/observed.metrics"
