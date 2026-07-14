# Hermes Agent adapter

Guarantee: **context-saving with a dedicated router-only Hermes profile; routing-only when the full catalog is still enabled**.

Hermes indexes installed and external skill directories into its prompt, while loading full skill content on demand. Its native profiles and skill opt-out controls can isolate a router-only catalog. See the current [Hermes Skills documentation](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills).

Create or select a dedicated profile that exposes only this router (and any intentional always-on skills). Do not configure the canonical Intermesh source roots as Hermes external skill directories: those directories would be indexed into the prompt and erase the metadata savings.

The generic `intermesh profile` command is available for symlink-only catalogs, but Hermes's native profile controls are preferred. Intermesh does not edit Hermes configuration in V0.
