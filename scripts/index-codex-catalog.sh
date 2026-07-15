#!/usr/bin/env bash
set -euo pipefail

HOME_ROOT=${HOME}
DATABASE=""
INTERMESH_BIN=${INTERMESH_BIN:-intermesh}

usage() {
    echo "usage: scripts/index-codex-catalog.sh [--home path] [--db path]" >&2
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --home)
            [[ $# -ge 2 ]] || { usage; exit 2; }
            HOME_ROOT=$2
            shift 2
            ;;
        --db)
            [[ $# -ge 2 ]] || { usage; exit 2; }
            DATABASE=$2
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

if [[ "$INTERMESH_BIN" == */* ]]; then
    [[ -x "$INTERMESH_BIN" ]] || { echo "intermesh binary is not executable: $INTERMESH_BIN" >&2; exit 1; }
elif ! command -v "$INTERMESH_BIN" >/dev/null 2>&1; then
    echo "intermesh binary not found: $INTERMESH_BIN" >&2
    exit 1
fi

roots=()
codex_root="$HOME_ROOT/.codex"
if [[ -d "$codex_root" ]]; then
    while IFS= read -r skill; do
        rel=${skill#"$codex_root/"}
        plugin=${rel%%/*}
        roots+=(--root "$plugin=$(dirname "$skill")")
    done < <(find "$codex_root" -mindepth 3 -maxdepth 4 -type f \
        \( -path '*/skills/SKILL.md' -o -path '*/skills/*/SKILL.md' \) \
        -print | LC_ALL=C sort)
fi

cache_root="$codex_root/plugins/cache"
if [[ -d "$cache_root" ]]; then
    while IFS= read -r skill; do
        rel=${skill#"$cache_root/"}
        rest=${rel#*/}
        plugin=${rest%%/*}
        roots+=(--root "$plugin=$(dirname "$skill")")
    done < <(find "$cache_root" -type f -path '*/skills/*/SKILL.md' -print | LC_ALL=C sort)
fi

local_root="$HOME_ROOT/projects/dotfiles/common/.codex/skills"
if [[ -d "$local_root" ]]; then
    while IFS= read -r skill; do
        namespace=local
        if [[ "$skill" == */.system/* ]]; then
            namespace=codex-system
        fi
        roots+=(--root "$namespace=$(dirname "$skill")")
    done < <(find "$local_root" -type f -name SKILL.md -print | LC_ALL=C sort)
fi

if [[ ${#roots[@]} -eq 0 ]]; then
    echo "no SKILL.md files found under $HOME_ROOT/.codex or the local dotfiles catalog" >&2
    exit 1
fi

args=(index "${roots[@]}")
if [[ -n "$DATABASE" ]]; then
    args+=(--db "$DATABASE")
fi
args+=(--json)

echo "indexing $((${#roots[@]} / 2)) Codex skill roots" >&2
exec "$INTERMESH_BIN" "${args[@]}"
