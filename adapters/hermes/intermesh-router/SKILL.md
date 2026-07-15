---
name: intermesh-router
description: Route requests through the external Intermesh registry and load only selected skills. Use when a Hermes profile exposes this router instead of the full catalog.
---

# Intermesh Router for Hermes Agent

For each substantive user request:

1. Route only the substantive task, not the router invocation wrapper. Form `$SUBSTANTIVE_REQUEST` by removing phrases whose sole purpose is to request Intermesh routing, permit the private route receipt, or constrain the routing procedure itself. Preserve every constraint that affects the requested task, its output, or skill applicability. Run `intermesh route --query "$SUBSTANTIVE_REQUEST" --host hermes --cwd "$PWD" --limit 3 --json`, adding known `--extension` and `--environment` gates.
2. Retain resolver `warnings` through applicability selection. Surface non-conflict warnings immediately; conflict warnings are potential conflicts until selection is complete.
3. Treat ranked candidates as retrieval results, not automatically selected skills. Use each candidate's `id`, `description`, and `reasons` to judge whether its documented trigger applies. Trigger boundaries and explicit exclusions outweigh loose keyword overlap; mentioning a file or tool as an object does not select its workflow. Preserve recall when applicability is genuinely ambiguous.
4. Starting from applicable ranked candidates, compute the `required_by` dependency closure: repeatedly include any candidate whose `required_by` contains an already selected candidate, even when that dependency also has `selected_by: "rank"`. Then evaluate each selected candidate's `conflicts_with` intersection with the selected set and surface only conflicts whose two endpoints are selected. Read the complete `skill_md` for that final set in returned array order. If none apply, continue without a skill and report the routing abstention.
5. Apply only selected instructions through Hermes's normal skill behavior; Intermesh supplies bounded discovery metadata and the agent owns applicability judgment.
6. Preserve the default hashed receipt and attach an Interspect outcome when that integration is available.

If routing reports an empty or unhealthy registry, run `intermesh doctor` and report the remediation. Never add the canonical source catalog as an extra Hermes skill directory merely to make routing work; that would preload its metadata again.
