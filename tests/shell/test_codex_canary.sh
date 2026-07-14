#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
TMP=$(mktemp -d "${TMPDIR:-/tmp}/intermesh-codex-canary-tests.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

assert_json() {
    local path=$1 expression=$2
    python3 - "$path" "$expression" <<'PY'
import json, sys
path, expression = sys.argv[1:]
with open(path) as handle:
    value = json.load(handle)
if not eval(expression, {"__builtins__": {}}, {"value": value}):
    raise SystemExit(f"assertion failed: {expression}\n{value!r}")
PY
}

source_home="$TMP/source-home"
normal_home="$TMP/normal-home"
canary_root="$TMP/canary"
workspace="$TMP/workspace"
router="$TMP/router"
mkdir -p "$source_home" "$normal_home/.codex" "$workspace" "$router"
printf '%s\n' 'normal-profile-sentinel' > "$normal_home/.codex/config.toml"
printf '%s\n' '---' 'name: intermesh-router' 'description: Route through Intermesh.' '---' > "$router/SKILL.md"

index_capture="$TMP/index-args.txt"
fake_index="$TMP/fake-index.sh"
cat > "$fake_index" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" > "$INDEX_CAPTURE"
database=""
while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--db" ]]; then
        database=$2
        shift 2
    else
        shift
    fi
done
[[ -n "$database" ]]
mkdir -p "$(dirname "$database")"
: > "$database"
printf '%s\n' '{"skill_count":170,"diagnostic_count":0,"fingerprint":"sha256:test"}'
SH
chmod +x "$fake_index"

baseline_prompt="$TMP/baseline-prompt.json"
canary_prompt="$TMP/canary-prompt.json"
cat > "$baseline_prompt" <<'JSON'
[
  {"type":"message","role":"developer","content":[
    {"type":"input_text","text":"<skills_instructions>\n## Skills\n### Skill roots\n- `r2` = `CANARY_CODEX/skills/.system`\n### Available skills\n- imagegen: Built-in image generation. (file: r2/imagegen/SKILL.md)\n- intertest:test-driven-development: A deliberately long baseline skill description that represents one of many automatically exposed plugin skills and makes the removable metadata budget measurable across a realistic catalog containing dozens of independent workflows and enough descriptive routing language to consume meaningful model context before any selected skill body is loaded. (file: /catalog/intertest/tdd/SKILL.md)\n- intertest:verification-before-completion: Another deliberately long baseline description for a second automatically exposed plugin skill in the normal Codex profile, including trigger boundaries, workflow intent, verification expectations, and enough additional routing detail to model the repeated per-skill cost found in a large installed plugin collection. (file: /catalog/intertest/verify/SKILL.md)\n- interskill:audit: Audit skill structure, routing metadata, progressive disclosure, references, scripts, portability, safety boundaries, and validation evidence whenever an existing reusable agent workflow needs a systematic quality review before publication or installation. (file: /catalog/interskill/audit/SKILL.md)\n</skills_instructions>"}
  ]}
]
JSON
cat > "$canary_prompt" <<'JSON'
[
  {"type":"message","role":"developer","content":[
    {"type":"input_text","text":"<skills_instructions>\n## Skills\n### Available skills\n- imagegen: Built-in image generation. (file: CANARY_CODEX/skills/.system/imagegen/SKILL.md)\n- intermesh:intermesh-router: Route through Intermesh. (file: ROUTER_PATH/SKILL.md)\n</skills_instructions>"}
  ]}
]
JSON
sed -i.bak "s|CANARY_CODEX|$canary_root/codex|g; s|ROUTER_PATH|$router|g" "$baseline_prompt" "$canary_prompt"
rm -f "$baseline_prompt.bak" "$canary_prompt.bak"

fake_intermesh="$TMP/fake-intermesh"
cat > "$fake_intermesh" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "doctor" ]]; then
    printf '%s\n' '{"version":1,"healthy":true,"skill_count":170,"diagnostic_count":0,"fingerprint":"sha256:test"}'
    exit 0
fi
echo "unexpected fake intermesh invocation: $*" >&2
exit 2
SH
chmod +x "$fake_intermesh"

fake_codex="$TMP/fake-codex"
cat > "$fake_codex" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "debug" && "${2:-}" == "prompt-input" ]]; then
    if [[ "${CODEX_HOME:-}" == "$CANARY_CODEX_HOME" ]]; then
        cat "$CANARY_PROMPT"
    else
        cat "$BASELINE_PROMPT"
    fi
    exit 0
fi
if [[ "${1:-}" == "exec" ]]; then
    mkdir -p "$XDG_STATE_HOME/intermesh"
    printf '%s\n' '{"event":"intermesh.route.v1","route_id":"route-test","candidates":[{"id":"intertest:verification-before-completion","skill_md":"/catalog/verify/SKILL.md"}],"warnings":[],"latency_micros":1200}' >> "$XDG_STATE_HOME/intermesh/routes.jsonl"
    printf '%s\n' \
        '{"type":"thread.started","thread_id":"thread-test"}' \
        '{"type":"turn.started"}' \
        '{"type":"item.completed","item":{"type":"agent_message","text":"ok"}}' \
        '{"type":"turn.completed","usage":{"input_tokens":1200,"cached_input_tokens":900,"output_tokens":20,"reasoning_output_tokens":5}}'
    exit 0
fi
if [[ "${1:-}" == "login" ]]; then
    printf '%s\n' "login-home=$HOME" "login-codex-home=$CODEX_HOME"
    exit 0
fi
echo "unexpected fake codex invocation: $*" >&2
exit 2
SH
chmod +x "$fake_codex"

common_env=(
    HOME="$normal_home"
    INTERMESH_CANARY_ROOT="$canary_root"
    INTERMESH_SOURCE_HOME="$source_home"
    INTERMESH_ROUTER_DIR="$router"
    INTERMESH_INDEX_SCRIPT="$fake_index"
    INTERMESH_BIN="$fake_intermesh"
    CODEX_BIN="$fake_codex"
    INDEX_CAPTURE="$index_capture"
    BASELINE_PROMPT="$baseline_prompt"
    CANARY_PROMPT="$canary_prompt"
    CANARY_CODEX_HOME="$canary_root/codex"
)

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" setup > "$TMP/setup.json"
assert_json "$TMP/setup.json" 'value["status"] == "ready" and value["source_skill_count"] == 170'
[[ -L "$canary_root/home/.agents/skills/intermesh-router" ]] || fail "router symlink missing"
[[ "$(readlink "$canary_root/home/.agents/skills/intermesh-router")" == "$router" ]] || fail "router symlink target changed"
grep -Fqx 'cli_auth_credentials_store = "file"' "$canary_root/codex/config.toml" || fail "isolated credential storage missing"
grep -Fq "HOME = \"$source_home\"" "$canary_root/codex/config.toml" || fail "task HOME restoration missing"
if grep -Fq 'PATH =' "$canary_root/codex/config.toml"; then
    fail "canary config must not freeze the setup PATH"
fi
grep -Fqx 'normal-profile-sentinel' "$normal_home/.codex/config.toml" || fail "normal profile was mutated"
grep -Fqx -- '--home' "$index_capture" || fail "source home flag missing"
grep -Fqx -- "$source_home" "$index_capture" || fail "source home value missing"

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" compare --workspace "$workspace" > "$TMP/context.json"
assert_json "$TMP/context.json" 'value["canary"]["non_system_skill_count"] == 1'
assert_json "$TMP/context.json" 'value["canary"]["unexpected_non_system_skills"] == []'
assert_json "$TMP/context.json" 'value["baseline"]["system_skill_count"] == 1 and value["baseline"]["non_system_skill_count"] == 3'
assert_json "$TMP/context.json" 'value["non_system_metadata_reduction"] > 0.8'
assert_json "$TMP/context.json" '"skills" not in value["baseline"] and "skills" not in value["canary"]'
[[ -f "$canary_root/results/context.json" ]] || fail "context comparison was not retained"

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" doctor --workspace "$workspace" > "$TMP/doctor.json"
assert_json "$TMP/doctor.json" 'value["healthy"] is True and value["unexpected_non_system_skills"] == []'

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" report > "$TMP/empty-report.json"
assert_json "$TMP/empty-report.json" 'value["verdict"] == "hold" and value["sessions"] == 0'

for index in $(seq 1 20); do
    session=$(printf 'case-%02d' "$index")
    env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" exec \
        --session "$session" --workspace "$workspace" \
        --expected intertest:verification-before-completion \
        -- "verify the change" > "$TMP/$session.events"
    env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" label \
        --session "$session" --task pass --routing correct \
        --note "synthetic shell test" > "$TMP/$session.label.json"
done

python3 - "$canary_root/results/sessions.jsonl" <<'PY'
import json, sys
rows = [json.loads(line) for line in open(sys.argv[1]) if line.strip()]
assert len(rows) == 20, rows
assert rows[-1]["usage"]["input_tokens"] == 1200, rows[-1]
assert rows[-1]["route_count"] == 1, rows[-1]
assert rows[-1]["candidate_ids"] == ["intertest:verification-before-completion"], rows[-1]
assert rows[-1]["expected_in_candidates"] is True, rows[-1]
PY

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" report > "$TMP/report-20.json"
assert_json "$TMP/report-20.json" 'value["sessions"] == 20 and value["labeled_sessions"] == 20'
assert_json "$TMP/report-20.json" 'value["gates"]["activation_ready"] is False'

for index in $(seq 21 30); do
    session=$(printf 'case-%02d' "$index")
    env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" exec \
        --session "$session" --workspace "$workspace" \
        --expected intertest:verification-before-completion \
        -- "verify the change" > "$TMP/$session.events"
    env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" label \
        --session "$session" --task pass --routing correct \
        --note "synthetic shell test" > "$TMP/$session.label.json"
done

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" report --check-gates > "$TMP/report-30.json"
assert_json "$TMP/report-30.json" 'value["sessions"] == 30 and value["labeled_sessions"] == 30'
assert_json "$TMP/report-30.json" 'value["gates"]["activation_ready"] is True'
assert_json "$TMP/report-30.json" 'value["usage"]["input_tokens_p50"] == 1200'

env "${common_env[@]}" "$ROOT/scripts/codex-canary.sh" uninstall --yes > "$TMP/uninstall.json"
assert_json "$TMP/uninstall.json" 'value["status"] == "removed"'
[[ ! -e "$canary_root" ]] || fail "canary root was not removed"
grep -Fqx 'normal-profile-sentinel' "$normal_home/.codex/config.toml" || fail "normal profile changed during uninstall"

echo "codex canary tests: PASS"
