# Claude Code adapter

Guarantee: **context-saving in a router-only installation; routing-only when other skill plugins remain installed**.

Claude's Agent Skills model preloads each available skill's name and description and loads the full instructions on demand. The Intermesh plugin therefore declares only `skills/router`; canonical source skills must remain outside Claude Code's discovered skill/plugin paths. See Anthropic's current [Agent Skills overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview).

Installing this plugin does not suppress skills contributed by other installed plugins. In that common configuration Intermesh improves routing but cannot claim metadata-context reduction. A context-saving experiment needs an isolated Claude Code plugin/profile containing this router alone (plus explicitly chosen always-on skills).

`intermesh profile` may manage a user-selected Claude skill catalog only when its removable entries are symlinks. It never edits Claude settings or uninstalls plugins.
