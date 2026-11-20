-- Named collections group catalogue entries for operator organization only.
-- Membership never enables, installs, or grants execution permissions to a tool.
CREATE TABLE IF NOT EXISTS tool_collections (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tool_collection_items (
    collection_id BIGINT NOT NULL REFERENCES tool_collections(id) ON DELETE CASCADE,
    tool_id       TEXT NOT NULL,
    position      INTEGER NOT NULL DEFAULT 0,
    added_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (collection_id, tool_id)
);

CREATE INDEX IF NOT EXISTS tool_collection_items_order_idx
    ON tool_collection_items (collection_id, position, tool_id);
