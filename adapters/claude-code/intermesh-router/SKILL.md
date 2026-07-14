---
name: intermesh-router
description: Route requests through the external Intermesh registry and load only its selected Agent Skills. Use when the router-only Intermesh profile is active in Claude Code.
allowed-tools: Bash, Read
---

# Intermesh Router for Claude Code

For each substantive user request:

1. Run `intermesh route --query "$USER_REQUEST" --host claude-code --cwd "$PWD" --limit 3 --json`. Include relevant `--extension` or `--environment` values when known.
2. Surface every resolver warning.
3. Use `Read` to load every returned `candidates[].skill_md` completely and in array order. Required skills precede their dependents.
4. Follow the selected skills normally; Intermesh is discovery infrastructure, not an instruction executor.
5. Keep the default hashed route receipt. When Interspect outcome attachment is present, connect the final outcome to that route.

If the registry is unavailable, run `intermesh doctor` and report the problem. Do not silently fall back to scanning an unbounded skill catalog into context.
