PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS registry_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS skills (
    id TEXT PRIMARY KEY,
    namespace TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    skill_md TEXT NOT NULL UNIQUE,
    directory TEXT NOT NULL,
    body_hash TEXT NOT NULL,
    manifest_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS triggers (
    skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (skill_id, kind, value)
);

CREATE TABLE IF NOT EXISTS edges (
    source_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    relation TEXT NOT NULL,
    target_id TEXT NOT NULL,
    PRIMARY KEY (source_id, relation, target_id)
);

CREATE TABLE IF NOT EXISTS roots (
    path TEXT NOT NULL,
    namespace TEXT NOT NULL,
    PRIMARY KEY (path, namespace)
);

CREATE TABLE IF NOT EXISTS diagnostics (
    path TEXT NOT NULL,
    code TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL,
    PRIMARY KEY (path, code, message)
);

CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name);
CREATE INDEX IF NOT EXISTS idx_triggers_lookup ON triggers(kind, value);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target_id, relation);
