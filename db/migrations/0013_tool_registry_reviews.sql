ALTER TABLE tool_registry
    ADD COLUMN IF NOT EXISTS review_status TEXT NOT NULL DEFAULT 'pending'
        CHECK (review_status IN ('pending', 'approved', 'rejected')),
    ADD COLUMN IF NOT EXISTS review_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reviewed_by TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS tool_registry_review_status_idx
    ON tool_registry (review_status, id);

-- Immutable operator decision history. Review decisions are metadata only:
-- they never change manifest permissions, enabled state, or execution handlers.
CREATE TABLE IF NOT EXISTS tool_registry_reviews (
    id BIGSERIAL PRIMARY KEY,
    registry_id TEXT NOT NULL REFERENCES tool_registry(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    note TEXT NOT NULL DEFAULT '',
    reviewed_by TEXT NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tool_registry_reviews_registry_idx
    ON tool_registry_reviews (registry_id, reviewed_at DESC, id DESC);
