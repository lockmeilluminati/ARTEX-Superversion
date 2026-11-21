-- Preserve provenance rows when a task is removed. Claims and runs remain as
-- history; the task association becomes unavailable rather than blocking deletion.
ALTER TABLE collection_runs ALTER COLUMN task_id DROP NOT NULL;
ALTER TABLE observation_claims ALTER COLUMN task_id DROP NOT NULL;
