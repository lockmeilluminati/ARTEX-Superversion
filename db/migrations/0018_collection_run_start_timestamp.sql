-- A planned collection run has not started a process yet. The timestamp is
-- stamped atomically by StartCollectionRun at the actual launch boundary.
ALTER TABLE collection_runs ALTER COLUMN started_at DROP NOT NULL;
ALTER TABLE collection_runs ALTER COLUMN started_at DROP DEFAULT;
