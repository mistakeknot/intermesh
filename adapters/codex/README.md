# Codex adapter

Guarantee: **context-saving only with an activated managed catalog; otherwise routing-only**.

Codex discovers skill metadata from configured skill roots and loads full `SKILL.md` bodies only when selected. Putting this router alone in an active root removes the other skills' discovery metadata from the prompt while leaving their canonical files available to Intermesh. See the current [Codex skill-creator guidance](https://github.com/openai/codex/blob/main/codex-rs/skills/src/assets/samples/skill-creator/SKILL.md).

The recommended trial is `scripts/codex-canary.sh setup`. It creates separate
`HOME` and `CODEX_HOME` roots, exposes this router as the only non-system skill,
indexes the canonical catalog out of band, and leaves the normal Codex profile
untouched. Codex's bundled system skills remain visible. Run `compare` and
`doctor` before starting sessions; see the [canary guide](../../docs/guides/codex-router-canary.md).

`intermesh profile plan` remains available for hosts whose catalog is already
a dedicated symlink-only root. V0 permits activation only when every removable
entry is a symlink; regular directories make the plan `routing_only`.
`profile apply --yes` snapshots the exact links before replacing them with
`intermesh-router`, and `profile restore --yes` reverses the operation.

The context-saving claim applies only to the catalog actually managed by the profile. Skills exposed through another configured root or plugin remain visible to Codex.

Route results include canonical candidate descriptions. The router treats the
top three as bounded retrieval results, applies the descriptions' trigger and
exclusion criteria, and loads complete `SKILL.md` bodies only for applicable
candidates and their requirements. This second progressive-disclosure gate
preserves top-three retrieval recall without paying every candidate-body cost.
