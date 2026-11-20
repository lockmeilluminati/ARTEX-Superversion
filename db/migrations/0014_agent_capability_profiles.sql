-- Persist observed runtime capabilities per agent. Profiles are descriptive
-- snapshots derived from tools actually assembled for that agent; they do not
-- grant execution permissions.
CREATE TABLE IF NOT EXISTS agent_capability_profiles (
    agent_key   TEXT PRIMARY KEY,
    capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    platforms    JSONB NOT NULL DEFAULT '[]'::jsonb,
    evidence     JSONB NOT NULL DEFAULT '[]'::jsonb,
    observed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_agent_capability_profiles_observed
    ON agent_capability_profiles(observed_at DESC);
