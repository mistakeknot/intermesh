---
name: router
description: Route requests through Intermesh's external registry and load only the selected Agent Skills. Use when this router-only plugin replaces a full in-prompt skill catalog.
allowed-tools: Bash, Read
---

# Intermesh Router for Claude Code

For each substantive user request:

1. Route only the substantive task, not the router invocation wrapper. Form `$SUBSTANTIVE_REQUEST` by removing phrases whose sole purpose is to request Intermesh routing, permit the private route receipt, or constrain the routing procedure itself. Preserve every constraint that affects the requested task, its output, or skill applicability. Run `intermesh route --query "$SUBSTANTIVE_REQUEST" --host claude-code --cwd "$PWD" --limit 3 --json`. Include relevant `--extension` or `--environment` values when known.
2. Retain resolver `warnings` through applicability selection. Surface non-conflict warnings immediately; conflict warnings are potential conflicts until selection is complete.
3. Treat ranked candidates as retrieval results, not automatically selected skills. Use each candidate's `id`, `description`, and `reasons` to judge whether its documented trigger applies. Trigger boundaries and explicit exclusions outweigh loose keyword overlap; mentioning a file or tool as an object does not select its workflow. Preserve recall when applicability is genuinely ambiguous.
4. Starting from applicable ranked candidates, compute the `required_by` dependency closure: repeatedly include any candidate whose `required_by` contains an already selected candidate, even when that dependency also has `selected_by: "rank"`. Then evaluate each selected candidate's `conflicts_with` intersection with the selected set and surface only conflicts whose two endpoints are selected. Read the complete `skill_md` for that final set in returned array order. If none apply, continue without a skill and report the routing abstention.
5. Follow the selected skills normally; Intermesh supplies bounded discovery metadata and leaves applicability judgment to the agent.
6. Keep the default hashed route receipt. When Interspect outcome attachment is present, connect the final outcome to that route.

If the registry is unavailable, run `intermesh doctor` and report the problem. Do not silently fall back to scanning an unbounded skill catalog into context.
