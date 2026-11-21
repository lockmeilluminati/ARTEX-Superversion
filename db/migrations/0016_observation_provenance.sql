-- Durable collection execution and append-only observation provenance.
-- Evidence rows are metadata for blobs already registered by the evidence layer;
-- this migration does not pretend that a locator or caller digest is a stored blob.
CREATE TABLE IF NOT EXISTS collection_runs (
    id                  UUID PRIMARY KEY,
    task_id             BIGINT NOT NULL REFERENCES tasks(id) ON DELETE SET NULL,
    exploration_id      BIGINT REFERENCES explorations(id) ON DELETE SET NULL,
    source_tool_id      TEXT NOT NULL,
    source_tool_version TEXT,
    execution_mode      TEXT NOT NULL DEFAULT 'unknown'
                        CHECK (execution_mode IN ('unknown','cli','gui','api','mcp','hybrid')),
    platform            TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'planned'
                        CHECK (status IN ('planned','running','succeeded','partial','failed','cancelled')),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at         TIMESTAMPTZ,
    exit_code           INTEGER,
    output_locator      TEXT NOT NULL DEFAULT '',
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (finished_at IS NULL OR finished_at >= started_at)
);
CREATE INDEX IF NOT EXISTS collection_runs_task_started_idx
    ON collection_runs(task_id, started_at DESC);
CREATE INDEX IF NOT EXISTS collection_runs_tool_started_idx
    ON collection_runs(source_tool_id, started_at DESC);

-- The evidence service must write and verify the content-addressed blob before
-- inserting this metadata row. storage_key is opaque; filesystem paths are not API.
CREATE TABLE IF NOT EXISTS evidence_artifacts (
    id              UUID PRIMARY KEY,
    sha256          CHAR(64) NOT NULL,
    byte_length     BIGINT NOT NULL CHECK (byte_length >= 0),
    media_type      TEXT NOT NULL DEFAULT 'application/octet-stream',
    storage_key     TEXT NOT NULL UNIQUE,
    task_id         BIGINT REFERENCES tasks(id) ON DELETE SET NULL,
    collection_run_id UUID REFERENCES collection_runs(id) ON DELETE SET NULL,
    source_tool_id  TEXT NOT NULL DEFAULT '',
    original_name   TEXT NOT NULL DEFAULT '',
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (sha256 ~ '^[0-9a-f]{64}$')
);
CREATE INDEX IF NOT EXISTS evidence_artifacts_digest_idx ON evidence_artifacts(sha256);
CREATE INDEX IF NOT EXISTS evidence_artifacts_run_idx ON evidence_artifacts(collection_run_id);

CREATE TABLE IF NOT EXISTS observation_claims (
    observation_id      TEXT PRIMARY KEY,
    identity_key        TEXT NOT NULL,
    observation_type    TEXT NOT NULL
                        CHECK (observation_type IN ('domain','host','ip','port','service','url','technology','finding')),
    target              TEXT NOT NULL,
    port                INTEGER CHECK (port BETWEEN 1 AND 65535),
    protocol            TEXT NOT NULL DEFAULT '',
    name                TEXT NOT NULL DEFAULT '',
    value               TEXT NOT NULL DEFAULT '',
    source_tool_id      TEXT NOT NULL,
    source_tool_version TEXT,
    source_run_id       UUID NOT NULL REFERENCES collection_runs(id) ON DELETE RESTRICT,
    task_id             BIGINT NOT NULL REFERENCES tasks(id) ON DELETE SET NULL,
    observed_at         TIMESTAMPTZ NOT NULL,
    evidence_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE RESTRICT,
    evidence_sha256     CHAR(64) NOT NULL CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    scope               TEXT NOT NULL CHECK (scope IN ('unknown','in_scope','out_of_scope')),
    confidence          TEXT NOT NULL CHECK (confidence IN ('unknown','inferred','low','medium','high')),
    verification        TEXT NOT NULL CHECK (verification IN ('not_assessed','unverified','verified','false_positive','blocked')),
    collection_status   TEXT NOT NULL CHECK (collection_status IN ('unknown','not_tested','tested_positive','tested_negative','partial','failed')),
    attributes          JSONB NOT NULL DEFAULT '{}'::jsonb,
    payload             JSONB NOT NULL,
    asset_id            BIGINT REFERENCES assets(id) ON DELETE SET NULL,
    finding_id          BIGINT REFERENCES findings(id) ON DELETE SET NULL,
    exploration_node_id BIGINT REFERENCES exploration_nodes(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- identity_key intentionally is not unique: conflicting source claims coexist.
CREATE INDEX IF NOT EXISTS observation_claims_identity_idx ON observation_claims(identity_key, observed_at DESC);
CREATE INDEX IF NOT EXISTS observation_claims_task_idx ON observation_claims(task_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS observation_claims_run_idx ON observation_claims(source_run_id);
CREATE INDEX IF NOT EXISTS observation_claims_asset_idx ON observation_claims(asset_id) WHERE asset_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS observation_claims_finding_idx ON observation_claims(finding_id) WHERE finding_id IS NOT NULL;
