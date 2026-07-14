---
name: router
description: Route requests through Intermesh's external registry and load only the selected Agent Skills. Use when this router-only plugin replaces a full in-prompt skill catalog.
allowed-tools: Bash, Read
---

# Intermesh Router

For each substantive request:

1. Run `intermesh route --query "$USER_REQUEST" --host claude-code --cwd "$PWD" --limit 3 --json`.
2. Surface resolver warnings.
3. Read every returned `candidates[].skill_md` completely, in array order.
4. Follow each selected skill normally.
5. Keep the default hashed receipt; attach its outcome through Interspect when available.

If the registry is unhealthy, run `intermesh doctor` and report the remediation instead of silently scanning the full catalog.
