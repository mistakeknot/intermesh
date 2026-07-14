---
name: intermesh-router
description: Route requests through the external Intermesh registry and load only its selected Agent Skills. Use when the router-only Intermesh profile is active in Claude Code.
allowed-tools: Bash, Read
---

# Intermesh Router for Claude Code

For each substantive user request:

1. Run `intermesh route --query "$USER_REQUEST" --host claude-code --cwd "$PWD" --limit 3 --json`. Include relevant `--extension` or `--environment` values when known.
2. Surface every resolver warning.
3. Treat ranked candidates as retrieval results, not automatically selected skills. Use each candidate's `id`, `description`, and `reasons` to judge whether its documented trigger applies. Trigger boundaries and explicit exclusions outweigh loose keyword overlap; mentioning a file or tool as an object does not select its workflow. Preserve recall when applicability is genuinely ambiguous.
4. Read the complete `skill_md` only for candidates whose descriptions apply, plus their `selected_by: "requirement"` dependencies, in array order. If none apply, continue without a skill and report the routing abstention.
5. Follow the selected skills normally; Intermesh supplies bounded discovery metadata and leaves applicability judgment to the agent.
6. Keep the default hashed route receipt. When Interspect outcome attachment is present, connect the final outcome to that route.

If the registry is unavailable, run `intermesh doctor` and report the problem. Do not silently fall back to scanning an unbounded skill catalog into context.
