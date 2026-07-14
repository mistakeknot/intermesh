---
name: intermesh-router
description: Route requests through the external Intermesh registry and load only selected skills. Use when a Hermes profile exposes this router instead of the full catalog.
---

# Intermesh Router for Hermes Agent

For each substantive user request:

1. Run `intermesh route --query "$USER_REQUEST" --host hermes --cwd "$PWD" --json`, adding known `--extension` and `--environment` gates.
2. Surface all warnings from the resolver.
3. Read every returned `candidates[].skill_md` file fully and in order; dependencies are ordered before dependents.
4. Apply those instructions through Hermes's normal skill behavior.
5. Preserve the default hashed receipt and attach an Interspect outcome when that integration is available.

If routing reports an empty or unhealthy registry, run `intermesh doctor` and report the remediation. Never add the canonical source catalog as an extra Hermes skill directory merely to make routing work; that would preload its metadata again.
