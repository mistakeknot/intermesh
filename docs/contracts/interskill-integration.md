# Interskill Integration Contract

Status: V1

Interskill owns skill authoring and audit quality. Intermesh owns optional routing metadata and validation. Neither tool is required for the other tool's base workflow.

## Authoring

When the `intermesh` executable is present, Interskill may offer an adjacent `intermesh.yaml` for skills with explicit triggers or relationships. The canonical instructions remain in `SKILL.md`; the manifest may contain only the versioned fields defined by [registry-v1.md](registry-v1.md).

After creating or changing a manifest, Interskill runs:

```bash
intermesh manifest validate /absolute/path/to/intermesh.yaml
```

A nonzero exit is an authoring failure to fix before completion. When `command -v intermesh` fails, Interskill continues its existing workflow without a manifest unless the user explicitly requested one.

## Audit

If `intermesh.yaml` exists and the CLI is available, Interskill includes validator diagnostics in its normal PASS/WARN/FAIL report. A missing CLI is `SKIP`, not a finding. A missing manifest is not a finding unless Intermesh metadata was an explicit requirement.

## Compatibility

- Integration is capability-detected; there is no hard package dependency.
- Interskill does not write the SQLite registry or infer learned weights.
- Intermesh does not rewrite `SKILL.md` or score authoring quality.
- Unknown V1 manifest fields remain forward-compatible, but unsupported manifest versions fail validation.
