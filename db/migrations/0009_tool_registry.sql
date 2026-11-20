CREATE TABLE IF NOT EXISTS tool_registry (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    manifest    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tool_registry_name_idx ON tool_registry (name);
