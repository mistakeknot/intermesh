#!/usr/bin/env bash
set -euo pipefail
umask 077

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
NORMAL_HOME=${HOME}
CANARY_ROOT=${INTERMESH_CANARY_ROOT:-$NORMAL_HOME/.local/share/intermesh/codex-canary}
SOURCE_HOME=${INTERMESH_SOURCE_HOME:-$NORMAL_HOME}
ROUTER_DIR=${INTERMESH_ROUTER_DIR:-$REPO_ROOT/adapters/codex/intermesh-router}
INDEX_SCRIPT=${INTERMESH_INDEX_SCRIPT:-$REPO_ROOT/scripts/index-codex-catalog.sh}
ANALYZER=${INTERMESH_CANARY_ANALYZER:-$REPO_ROOT/scripts/codex-canary-analyze.py}
INTERMESH_BIN=${INTERMESH_BIN:-intermesh}
CODEX_BIN=${CODEX_BIN:-codex}

CANARY_HOME=$CANARY_ROOT/home
CODEX_HOME_DIR=$CANARY_ROOT/codex
STATE_HOME=$CANARY_ROOT/state
RESULTS_DIR=$CANARY_ROOT/results
RUNS_DIR=$CANARY_ROOT/runs
DATABASE=$CANARY_HOME/.local/share/intermesh/registry.db
ROUTER_LINK=$CANARY_HOME/.agents/skills/intermesh-router
MARKER=$CANARY_ROOT/.intermesh-codex-canary-v1

usage() {
    cat >&2 <<'EOF'
usage: scripts/codex-canary.sh <command> [options]

commands:
  setup
  inspect [--workspace path]
  compare [--workspace path]
  doctor [--workspace path]
  login [codex login arguments...]
  run [--workspace path] [-- codex arguments...]
  exec --session id [--workspace path] [--expected id[,id]] -- prompt
  label --session id --task pass|partial|fail --routing correct|partial|wrong|abstain [--note text]
  report [--check-gates]
  uninstall --yes

Environment overrides: INTERMESH_CANARY_ROOT, INTERMESH_SOURCE_HOME,
INTERMESH_ROUTER_DIR, INTERMESH_BIN, and CODEX_BIN.
EOF
}

die() {
    echo "$*" >&2
    exit 2
}

require_setup() {
    [[ -f "$MARKER" ]] || die "canary is not set up: run scripts/codex-canary.sh setup"
    [[ -d "$CODEX_HOME_DIR" ]] || die "canary CODEX_HOME is missing: $CODEX_HOME_DIR"
}

validate_root() {
    [[ "$CANARY_ROOT" == /* ]] || die "canary root must be absolute: $CANARY_ROOT"
    [[ "$CANARY_ROOT" != "/" ]] || die "refusing unsafe canary root: /"
    [[ "$CANARY_ROOT" != "$NORMAL_HOME" ]] || die "canary root must not be the normal HOME"
    [[ "$CANARY_ROOT" != "$SOURCE_HOME" ]] || die "canary root must not be the source HOME"
}

resolve_executable() {
    local executable=$1
    if [[ "$executable" == */* ]]; then
        [[ -x "$executable" ]] || die "executable not found: $executable"
        local directory
        directory=$(cd "$(dirname "$executable")" && pwd)
        printf '%s/%s\n' "$directory" "$(basename "$executable")"
    else
        command -v "$executable" || die "executable not found on PATH: $executable"
    fi
}

canary_env() {
    HOME="$CANARY_HOME" \
    CODEX_HOME="$CODEX_HOME_DIR" \
    CODEX_SQLITE_HOME="$CODEX_HOME_DIR" \
    XDG_STATE_HOME="$STATE_HOME" \
    PATH="$CANARY_ROOT/bin:$PATH" \
        "$@"
}

workspace_option() {
    local workspace=$PWD
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --workspace)
                [[ $# -ge 2 ]] || die "--workspace requires a path"
                workspace=$2
                shift 2
                ;;
            *)
                die "unexpected argument: $1"
                ;;
        esac
    done
    [[ -d "$workspace" ]] || die "workspace is not a directory: $workspace"
    (cd "$workspace" && pwd)
}

setup_canary() {
    validate_root
    [[ -f "$ROUTER_DIR/SKILL.md" ]] || die "router SKILL.md not found: $ROUTER_DIR/SKILL.md"
    [[ -x "$INDEX_SCRIPT" ]] || die "index helper is not executable: $INDEX_SCRIPT"
    [[ -x "$ANALYZER" ]] || die "canary analyzer is not executable: $ANALYZER"
    if [[ -e "$CANARY_ROOT" && ! -f "$MARKER" ]]; then
        die "refusing unmarked existing directory: $CANARY_ROOT"
    fi

    local real_intermesh
    real_intermesh=$(resolve_executable "$INTERMESH_BIN")
    mkdir -p \
        "$CANARY_HOME/.agents/skills" \
        "$(dirname "$DATABASE")" \
        "$CODEX_HOME_DIR" "$STATE_HOME" "$RESULTS_DIR" "$RUNS_DIR" "$CANARY_ROOT/bin"
    chmod 700 "$CANARY_ROOT" "$CANARY_HOME" "$CODEX_HOME_DIR" "$STATE_HOME" "$RESULTS_DIR" "$RUNS_DIR"

    if [[ -L "$ROUTER_LINK" ]]; then
        [[ "$(readlink "$ROUTER_LINK")" == "$ROUTER_DIR" ]] || die "router link points somewhere else: $ROUTER_LINK"
    elif [[ -e "$ROUTER_LINK" ]]; then
        die "router catalog entry is not a symlink: $ROUTER_LINK"
    else
        ln -s "$ROUTER_DIR" "$ROUTER_LINK"
    fi

    local wrapper_tmp=$CANARY_ROOT/bin/.intermesh.tmp
    {
        printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail'
        printf 'REAL_INTERMESH=%q\n' "$real_intermesh"
        printf 'CANARY_DATABASE=%q\n' "$DATABASE"
        cat <<'SH'
command=${1:-}
if [[ -z "$command" ]]; then
    exec "$REAL_INTERMESH"
fi
shift
case "$command" in
    route|search|resolve|graph|doctor)
        exec "$REAL_INTERMESH" "$command" --db "$CANARY_DATABASE" "$@"
        ;;
    *)
        exec "$REAL_INTERMESH" "$command" "$@"
        ;;
esac
SH
    } > "$wrapper_tmp"
    chmod 700 "$wrapper_tmp"
    mv "$wrapper_tmp" "$CANARY_ROOT/bin/intermesh"

    local config_tmp=$CODEX_HOME_DIR/.config.toml.tmp
    python3 - "$SOURCE_HOME" "$STATE_HOME" > "$config_tmp" <<'PY'
import json, sys
home, state = sys.argv[1:]
print('cli_auth_credentials_store = "file"')
print('mcp_oauth_credentials_store = "file"')
print('')
print('[shell_environment_policy]')
print('inherit = "all"')
print('set = { HOME = %s, XDG_STATE_HOME = %s }' % tuple(map(json.dumps, (home, state))))
PY
    chmod 600 "$config_tmp"
    mv "$config_tmp" "$CODEX_HOME_DIR/config.toml"

    local index_json=$RESULTS_DIR/index.json.tmp
    INTERMESH_BIN="$real_intermesh" "$INDEX_SCRIPT" \
        --home "$SOURCE_HOME" --db "$DATABASE" > "$index_json"
    mv "$index_json" "$RESULTS_DIR/index.json"

    {
        printf '%s\n' 'version=1'
        printf 'source_home=%s\n' "$SOURCE_HOME"
        printf 'router_dir=%s\n' "$ROUTER_DIR"
    } > "$MARKER"
    chmod 600 "$MARKER"

    python3 - "$CANARY_ROOT" "$SOURCE_HOME" "$ROUTER_DIR" "$RESULTS_DIR/index.json" <<'PY'
import json, sys
root, source, router, index_path = sys.argv[1:]
index = json.load(open(index_path))
print(json.dumps({
    "version": 1,
    "status": "ready",
    "canary_root": root,
    "source_home": source,
    "router": router,
    "source_skill_count": index.get("skill_count", 0),
    "source_diagnostic_count": index.get("diagnostic_count", 0),
}, indent=2, sort_keys=True))
PY
}

capture_prompt() {
    local mode=$1 workspace=$2 output=$3
    if [[ "$mode" == "canary" ]]; then
        (
            cd "$workspace"
            canary_env "$CODEX_BIN" debug prompt-input "Intermesh canary context inspection"
        ) > "$output"
    else
        (
            cd "$workspace"
            "$CODEX_BIN" debug prompt-input "Intermesh canary baseline inspection"
        ) > "$output"
    fi
}

inspect_canary() {
    require_setup
    local workspace
    workspace=$(workspace_option "$@")
    local prompt
    prompt=$(mktemp "${TMPDIR:-/tmp}/intermesh-canary-prompt.XXXXXX")
    trap 'rm -f "$prompt"' RETURN
    capture_prompt canary "$workspace" "$prompt"
    "$ANALYZER" prompt --input "$prompt" --router "$ROUTER_DIR" --workspace "$workspace"
    rm -f "$prompt"
    trap - RETURN
}

compare_canary() {
    require_setup
    local workspace
    workspace=$(workspace_option "$@")
    local baseline canary output
    baseline=$(mktemp "${TMPDIR:-/tmp}/intermesh-baseline-prompt.XXXXXX")
    canary=$(mktemp "${TMPDIR:-/tmp}/intermesh-canary-prompt.XXXXXX")
    output=$(mktemp "${TMPDIR:-/tmp}/intermesh-context.XXXXXX")
    trap 'rm -f "$baseline" "$canary" "$output"' RETURN
    capture_prompt baseline "$workspace" "$baseline"
    capture_prompt canary "$workspace" "$canary"
    "$ANALYZER" compare \
        --baseline "$baseline" --canary "$canary" \
        --router "$ROUTER_DIR" --workspace "$workspace" > "$output"
    cp "$output" "$RESULTS_DIR/context.json"
    chmod 600 "$RESULTS_DIR/context.json"
    cat "$output"
    rm -f "$baseline" "$canary" "$output"
    trap - RETURN
}

doctor_canary() {
    require_setup
    local workspace
    workspace=$(workspace_option "$@")
    [[ -L "$ROUTER_LINK" ]] || die "router link is missing: $ROUTER_LINK"
    [[ "$(readlink "$ROUTER_LINK")" == "$ROUTER_DIR" ]] || die "router link target drifted"
    grep -Fqx 'cli_auth_credentials_store = "file"' "$CODEX_HOME_DIR/config.toml" || die "isolated credential policy drifted"

    local registry context
    registry=$(mktemp "${TMPDIR:-/tmp}/intermesh-canary-doctor.XXXXXX")
    context=$(mktemp "${TMPDIR:-/tmp}/intermesh-canary-context.XXXXXX")
    trap 'rm -f "$registry" "$context"' RETURN
    "$CANARY_ROOT/bin/intermesh" doctor --json > "$registry"
    inspect_canary --workspace "$workspace" > "$context"
    python3 - "$registry" "$context" "$CANARY_ROOT" <<'PY'
import json, sys
registry = json.load(open(sys.argv[1]))
context = json.load(open(sys.argv[2]))
healthy = bool(registry.get("healthy")) and context.get("router_present") and not context.get("unexpected_non_system_skills")
print(json.dumps({
    "version": 1,
    "healthy": healthy,
    "canary_root": sys.argv[3],
    "registry": registry,
    "router_present": context.get("router_present", False),
    "system_skill_count": context.get("system_skill_count", 0),
    "unexpected_non_system_skills": context.get("unexpected_non_system_skills", []),
}, indent=2, sort_keys=True))
raise SystemExit(0 if healthy else 1)
PY
    rm -f "$registry" "$context"
    trap - RETURN
}

run_canary() {
    require_setup
    local workspace=$PWD
    if [[ "${1:-}" == "--workspace" ]]; then
        [[ $# -ge 2 ]] || die "--workspace requires a path"
        workspace=$2
        shift 2
    fi
    if [[ "${1:-}" == "--" ]]; then
        shift
    fi
    [[ -d "$workspace" ]] || die "workspace is not a directory: $workspace"
    (cd "$workspace" && canary_env "$CODEX_BIN" "$@")
}

exec_canary() {
    require_setup
    local session="" workspace=$PWD expected=""
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --session)
                [[ $# -ge 2 ]] || die "--session requires an ID"
                session=$2
                shift 2
                ;;
            --workspace)
                [[ $# -ge 2 ]] || die "--workspace requires a path"
                workspace=$2
                shift 2
                ;;
            --expected)
                [[ $# -ge 2 ]] || die "--expected requires a comma-separated skill ID"
                expected=$2
                shift 2
                ;;
            --)
                shift
                break
                ;;
            *)
                die "unexpected exec argument: $1"
                ;;
        esac
    done
    [[ "$session" =~ ^[A-Za-z0-9._-]+$ ]] || die "session ID must use letters, numbers, dot, underscore, or dash"
    [[ -d "$workspace" ]] || die "workspace is not a directory: $workspace"
    [[ $# -eq 1 ]] || die "exec requires exactly one prompt after --"
    local prompt=$1 run_dir=$RUNS_DIR/$session
    [[ ! -e "$run_dir" ]] || die "session already exists: $session"
    mkdir -p "$run_dir"
    chmod 700 "$run_dir"

    local route_log=$STATE_HOME/intermesh/routes.jsonl
    local before=0
    if [[ -f "$route_log" ]]; then
        before=$(wc -l < "$route_log" | tr -d ' ')
    fi
    local events=$run_dir/events.jsonl routes=$run_dir/routes.jsonl summary=$run_dir/summary.json
    : > "$routes"
    chmod 600 "$routes"

    set +e
    (
        cd "$workspace"
        canary_env "$CODEX_BIN" exec --ephemeral --json "$prompt"
    ) | tee "$events"
    local codex_status=${PIPESTATUS[0]}
    set -e
    chmod 600 "$events"

    if [[ -f "$route_log" ]]; then
        local after
        after=$(wc -l < "$route_log" | tr -d ' ')
        if (( after > before )); then
            tail -n "+$((before + 1))" "$route_log" > "$routes"
        fi
    fi

    "$ANALYZER" session \
        --events "$events" --routes "$routes" --session "$session" \
        --workspace "$workspace" --expected "$expected" --prompt "$prompt" \
        --exit-code "$codex_status" > "$summary"
    chmod 600 "$summary"
    mkdir -p "$RESULTS_DIR"
    python3 - "$summary" >> "$RESULTS_DIR/sessions.jsonl" <<'PY'
import json, sys
print(json.dumps(json.load(open(sys.argv[1])), separators=(",", ":"), sort_keys=True))
PY
    chmod 600 "$RESULTS_DIR/sessions.jsonl"
    return "$codex_status"
}

label_canary() {
    require_setup
    local session="" task="" routing="" note=""
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --session) session=${2:-}; shift 2 ;;
            --task) task=${2:-}; shift 2 ;;
            --routing) routing=${2:-}; shift 2 ;;
            --note) note=${2:-}; shift 2 ;;
            *) die "unexpected label argument: $1" ;;
        esac
    done
    [[ -n "$session" && -f "$RUNS_DIR/$session/summary.json" ]] || die "unknown session: $session"
    local label_tmp
    label_tmp=$(mktemp "${TMPDIR:-/tmp}/intermesh-canary-label.XXXXXX")
    "$ANALYZER" label --session "$session" --task "$task" --routing "$routing" --note "$note" > "$label_tmp"
    python3 - "$label_tmp" >> "$RESULTS_DIR/labels.jsonl" <<'PY'
import json, sys
print(json.dumps(json.load(open(sys.argv[1])), separators=(",", ":"), sort_keys=True))
PY
    chmod 600 "$RESULTS_DIR/labels.jsonl"
    cat "$label_tmp"
    rm -f "$label_tmp"
}

report_canary() {
    require_setup
    local check_gates=0
    if [[ "${1:-}" == "--check-gates" ]]; then
        check_gates=1
        shift
    fi
    [[ $# -eq 0 ]] || die "unexpected report argument: $1"
    local arguments=(
        report
        --sessions "$RESULTS_DIR/sessions.jsonl"
        --labels "$RESULTS_DIR/labels.jsonl"
        --context "$RESULTS_DIR/context.json"
    )
    if (( check_gates )); then
        arguments+=(--check-gates)
    fi
    "$ANALYZER" "${arguments[@]}"
}

uninstall_canary() {
    validate_root
    [[ "${1:-}" == "--yes" && $# -eq 1 ]] || die "uninstall requires --yes"
    require_setup
    grep -Fqx 'version=1' "$MARKER" || die "canary marker is invalid; refusing removal"
    rm -rf -- "$CANARY_ROOT"
    printf '{"status":"removed","canary_root":"%s"}\n' "$CANARY_ROOT"
}

command=${1:-}
[[ -n "$command" ]] || { usage; exit 2; }
shift
case "$command" in
    setup) [[ $# -eq 0 ]] || die "setup takes no arguments"; setup_canary ;;
    inspect) inspect_canary "$@" ;;
    compare) compare_canary "$@" ;;
    doctor) doctor_canary "$@" ;;
    login) require_setup; canary_env "$CODEX_BIN" login "$@" ;;
    run) run_canary "$@" ;;
    exec) exec_canary "$@" ;;
    label) label_canary "$@" ;;
    report) report_canary "$@" ;;
    uninstall) uninstall_canary "$@" ;;
    -h|--help|help) usage ;;
    *) usage; die "unknown command: $command" ;;
esac
