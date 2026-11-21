-- Preserve failed collection attempts as explicit coverage outcomes.
-- Freshness is derived from updated_at at read time; no timestamp is rewritten.
ALTER TABLE coverage_ledger
    DROP CONSTRAINT IF EXISTS coverage_ledger_outcome_check;
ALTER TABLE coverage_ledger
    ADD CONSTRAINT coverage_ledger_outcome_check
    CHECK (outcome IN ('reported','no_issue_found','ruled_out','not_applicable','needs_follow_up','test_failed'));
