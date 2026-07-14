#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
TMP=$(mktemp -d "${TMPDIR:-/tmp}/intermesh-setup-tests.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

assert_contains() {
    local file=$1 expected=$2
    grep -Fqx -- "$expected" "$file" || fail "$file does not contain argument: $expected"
}

fixture_home="$TMP/home"
mkdir -p \
    "$fixture_home/.codex/clavain/skills/debug" \
    "$fixture_home/.codex/plugins/cache/openai-curated-remote/notion/0.1.7/skills/capture" \
    "$fixture_home/projects/dotfiles/common/.codex/skills/argus" \
    "$fixture_home/projects/dotfiles/common/.codex/skills/.system/imagegen"
printf '%s\n' '---' 'name: debug' 'description: Debug.' '---' > "$fixture_home/.codex/clavain/skills/debug/SKILL.md"
printf '%s\n' '---' 'name: capture' 'description: Capture.' '---' > "$fixture_home/.codex/plugins/cache/openai-curated-remote/notion/0.1.7/skills/capture/SKILL.md"
printf '%s\n' '---' 'name: argus' 'description: Generate fixtures.' '---' > "$fixture_home/projects/dotfiles/common/.codex/skills/argus/SKILL.md"
printf '%s\n' '---' 'name: imagegen' 'description: Generate images.' '---' > "$fixture_home/projects/dotfiles/common/.codex/skills/.system/imagegen/SKILL.md"

capture="$TMP/index-args.txt"
fake_intermesh="$TMP/intermesh"
cat > "$fake_intermesh" <<'SH'
#!/usr/bin/env bash
printf '%s\n' "$@" > "$CAPTURE"
printf '%s\n' '{"skill_count":4,"diagnostic_count":0}'
SH
chmod +x "$fake_intermesh"

CAPTURE="$capture" INTERMESH_BIN="$fake_intermesh" \
    "$ROOT/scripts/index-codex-catalog.sh" \
    --home "$fixture_home" --db "$TMP/registry.db" > "$TMP/index.json"

assert_contains "$capture" index
assert_contains "$capture" --root
assert_contains "$capture" "clavain=$fixture_home/.codex/clavain/skills/debug"
assert_contains "$capture" "notion=$fixture_home/.codex/plugins/cache/openai-curated-remote/notion/0.1.7/skills/capture"
assert_contains "$capture" "local=$fixture_home/projects/dotfiles/common/.codex/skills/argus"
assert_contains "$capture" "codex-system=$fixture_home/projects/dotfiles/common/.codex/skills/.system/imagegen"
assert_contains "$capture" --db
assert_contains "$capture" "$TMP/registry.db"
assert_contains "$capture" --json

empty_home="$TMP/empty-home"
mkdir -p "$empty_home"
if CAPTURE="$capture" INTERMESH_BIN="$fake_intermesh" \
    "$ROOT/scripts/index-codex-catalog.sh" --home "$empty_home" \
    > "$TMP/empty.out" 2> "$TMP/empty.err"; then
    fail "empty catalog should be rejected"
fi
grep -Fq 'no SKILL.md files found' "$TMP/empty.err" || fail "empty-catalog diagnostic missing"

fake_index="$TMP/fake-index.sh"
cat > "$fake_index" <<'SH'
#!/usr/bin/env bash
exit 0
SH
chmod +x "$fake_index"

fake_eval="$TMP/intermesh-eval"
cat > "$fake_eval" <<'SH'
#!/usr/bin/env bash
cat "$EVAL_JSON"
SH
chmod +x "$fake_eval"

cat > "$TMP/eval-good.json" <<'JSON'
{
  "metrics": {
    "match_cases": 30,
    "no_match_cases": 30,
    "top_1_recall": 0.7333333333,
    "top_3_recall": 0.9666666667,
    "top_5_recall": 1.0,
    "mrr": 0.842,
    "no_match_precision": 1.0,
    "no_match_recall": 0.25
  },
  "warm_latency": {"p95_micros": 3200},
  "predictions": {}
}
JSON

EVAL_JSON="$TMP/eval-good.json" \
INTERMESH_BIN="$fake_intermesh" \
INTERMESH_EVAL_BIN="$fake_eval" \
INTERMESH_INDEX_SCRIPT="$fake_index" \
INTERMESH_BENCH_HOME="$fixture_home" \
INTERMESH_SKIP_TESTS=1 \
    "$ROOT/scripts/interlab-abstention.sh" > "$TMP/metrics.txt"

assert_contains "$TMP/metrics.txt" 'METRIC no_match_recall=0.25'
assert_contains "$TMP/metrics.txt" 'METRIC no_match_precision=1'
assert_contains "$TMP/metrics.txt" 'METRIC top3_recall=0.9666666667'
assert_contains "$TMP/metrics.txt" 'METRIC top5_recall=1'
assert_contains "$TMP/metrics.txt" 'METRIC mrr=0.842'
assert_contains "$TMP/metrics.txt" 'METRIC warm_p95_micros=3200'

python3 - "$TMP/eval-good.json" "$TMP/eval-bad.json" <<'PY'
import json, sys
source, target = sys.argv[1:]
data = json.load(open(source))
data["metrics"]["top_3_recall"] = 0.90
json.dump(data, open(target, "w"))
PY

if EVAL_JSON="$TMP/eval-bad.json" \
    INTERMESH_BIN="$fake_intermesh" \
    INTERMESH_EVAL_BIN="$fake_eval" \
    INTERMESH_INDEX_SCRIPT="$fake_index" \
    INTERMESH_BENCH_HOME="$fixture_home" \
    INTERMESH_SKIP_TESTS=1 \
        "$ROOT/scripts/interlab-abstention.sh" \
        > "$TMP/bad.out" 2> "$TMP/bad.err"; then
    fail "benchmark should reject a top-3 recall regression"
fi
grep -Fq 'top3_recall below 0.95' "$TMP/bad.err" || fail "constraint diagnostic missing"

echo "setup script tests: PASS"
