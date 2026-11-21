# Collection Coverage Intelligence

Status: additive extension of ARTEX's existing asset coverage and risk-class ledger. This does not replace the normal ARTEX task, planner, worker, asset, finding or evidence paths.

## Existing foundations reused

- Asset coverage denominator and untested-asset backlog: `db/task_assets_context.go`, based on declared task scope, direct source-task context and fact-anchored assets.
- Agent-reported risk-class coverage: `db/coverage_ledger.go`, written by the worker's `record_coverage` tool and surfaced in planner context/wrap-up.
- Existing exploration facts and findings remain the source of evidence and vulnerability records.
- Coverage can be disabled per task using the existing `coverage_enabled` setting; this change does not change that switch or the default.

## Coverage states

| State | Meaning |
|---|---|
| `not_tested` | No coverage entry exists for the requested surface and risk class. |
| `tested_positive` | A finding/report was recorded for the surface and risk class. |
| `tested_negative` | A negative result was explicitly recorded with evidence. |
| `not_applicable` | The class was explicitly ruled out or marked not applicable, with a reason. |
| `partial` | Work needs follow-up; it is not a clean result. |
| `failed` | The collection/test attempt failed; this is not a negative result. |
| `stale` | A prior assessment exists but is older than the caller's freshness window. The previous state is retained in the assessment. |
| `unknown` | The entry is missing required provenance (e.g. timestamp/evidence/reason) or has an unrecognized outcome. |

The `test_failed` ledger outcome is persisted in the existing coverage ledger and history. It is always an open gap. The pure `AssessCoverageEntry` helper classifies entries without mutating stored data. Callers choose the freshness window; ARTEX's ledger read uses a conservative 30-day default for its derived assessment. A future timestamp is treated as unknown rather than fresh.

## Important interpretation limits

- No finding is not a negative test.
- A missing risk-class entry is untested, not clean.
- A failed or partial attempt is not a negative result.
- A stale entry does not prove the surface changed; it means the old test should not be treated as current.
- Asset-level coverage currently uses fact anchors as its tested signal. That is a coarse "investigated" signal, not proof that every vulnerability class or discovery dimension was tested. Risk-class ledger entries remain the more specific signal.
- The coverage ledger is agent-reported. Evidence strings are references/notes, not yet foreign-key-enforced evidence claim IDs. Durable canonical observation claims remain a Job 9 dependency.

## Compatibility

This is an additive change to the existing ARTEX program. It does not change tool execution, target scope, approval, interception, egress or agent bindings. Existing ledger rows remain valid. The new migration only extends the allowed outcome set; `db/schema.sql` is updated for clean database creation.
