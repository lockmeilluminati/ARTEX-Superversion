-- Enforce the planned-versus-started timestamp invariant at the database
-- boundary as well as in CreateCollectionRun/StartCollectionRun.
UPDATE collection_runs SET started_at = NULL WHERE status = 'planned';

ALTER TABLE collection_runs
    ADD CONSTRAINT collection_runs_started_at_state_check
    CHECK (
        (status = 'planned' AND started_at IS NULL)
        OR
        (status <> 'planned' AND started_at IS NOT NULL)
    );
