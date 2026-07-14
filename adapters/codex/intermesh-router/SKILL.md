---
name: intermesh-router
description: Route each request through the external Intermesh skill registry, then load only the selected skill instructions. Use when Intermesh is the active Codex skill profile.
---

# Intermesh Router for Codex

For each substantive user request:

1. Run `intermesh route --query "$USER_REQUEST" --host codex --cwd "$PWD" --limit 3 --json`. Pass relevant `--extension` or `--environment` values when they are known.
2. Inspect `warnings` and tell the user about any conflict, missing requirement, or degraded routing result.
3. Read every returned `candidates[].skill_md` file completely, in array order. Required skills appear before the skill that required them.
4. Follow those skill instructions through Codex's normal skill mechanism. Intermesh selects instructions; it does not replace them.
5. Leave the default hashed routing receipt enabled. If Interspect outcome attachment is installed, associate the eventual outcome with the returned route receipt.

If routing fails because the registry is empty, run `intermesh doctor` and ask the user to rebuild it with `intermesh index`. Do not search or execute arbitrary files as a substitute for a failed route.
