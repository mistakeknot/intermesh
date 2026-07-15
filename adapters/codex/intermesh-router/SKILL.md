---
name: intermesh-router
description: Route each request through the external Intermesh skill registry, then load only the selected skill instructions. Use when Intermesh is the active Codex skill profile.
---

# Intermesh Router for Codex

For each substantive user request:

1. Route only the substantive task, not the router invocation wrapper. Form `$SUBSTANTIVE_REQUEST` by removing phrases whose sole purpose is to request Intermesh routing, permit the private route receipt, or constrain the routing procedure itself. Preserve every constraint that affects the requested task, its output, or skill applicability. Run `intermesh route --query "$SUBSTANTIVE_REQUEST" --host codex --cwd "$PWD" --limit 3 --json`. Pass relevant `--extension` or `--environment` values when they are known.
2. Retain resolver `warnings` through applicability selection. Surface non-conflict warnings immediately; conflict warnings are potential conflicts until selection is complete.
3. Treat ranked candidates as retrieval results, not automatically selected skills. Use each candidate's `id`, `description`, and `reasons` to decide whether its documented trigger applies to the request. Trigger boundaries and explicit exclusions in the description outweigh loose keyword overlap; mentioning a file or tool as an object is not by itself a request to operate on it. When genuinely ambiguous, preserve recall by selecting the candidate.
4. Starting from applicable ranked candidates, compute the `required_by` dependency closure: repeatedly include any candidate whose `required_by` contains an already selected candidate, even when that dependency also has `selected_by: "rank"`. Then evaluate each selected candidate's `conflicts_with` intersection with the selected set and surface only conflicts whose two endpoints are selected. Read the complete `skill_md` for that final set in returned array order. If no candidate applies, continue without a skill and report the routing abstention.
5. Follow selected skill instructions through Codex's normal skill mechanism. Intermesh supplies bounded discovery metadata; the agent retains applicability judgment.
6. Leave the default hashed routing receipt enabled. If Interspect outcome attachment is installed, associate the eventual outcome with the returned route receipt.

If routing fails because the registry is empty, run `intermesh doctor` and ask the user to rebuild it with `intermesh index`. Do not search or execute arbitrary files as a substitute for a failed route.
