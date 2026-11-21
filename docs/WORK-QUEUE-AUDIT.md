# ARTEX-Superversion Work Queue Audit

Reviewed: 2026-11-20

## Is the countdown complete?

No. The earlier statement that "10 substantive jobs remain" described the condensed next-up sequence, not the complete project backlog. It is not an accurate count of all remaining work. The roadmap has 16 broad workstreams plus cross-cutting validation, security, release and integration tasks. Several numbered items bundle multiple independent deliverables; some are ongoing maintenance rather than one-time jobs.

Use the project checklist as the operational punch list and the roadmap as the full capability map. Do not subtract an item merely because a design document or isolated package has been committed.

## Remaining intelligence work

1. Operationally verify the actual collection pipeline, scanner binaries, versions and command-to-output-to-ingestion path.
2. Complete canonical observation provenance and persistence. Committed migrations 0016–0017 and DB primitives now create collection-run rows, evidence metadata and append-only claims with exact-replay/collision handling. Added generic immutable artifact import through the existing evidence store, including bounded run-directory file validation, SHA-256/length calculation, metadata registration and GC retention; focused path/size tests are committed but unrun. Still missing: real process-boundary run-ID wiring, adapter integration, durable claim writes during ingestion, transactional projection links and executed PostgreSQL lifecycle tests.
3. Finish concurrency-safe idempotent ingestion, preserve conflicting claims and first/last-seen/source lineage, and test malformed, truncated, partial, empty and replay cases.
4. Integrate asset correlation with durable evidence claims and findings; add bounded DB loading and authenticated graph API/UI.
5. Complete collection coverage: operator API/UI, per-asset/per-dimension evidence, durable evidence references, freshness policy and tests.
6. Complete enrichment: feed evidence-backed projections into existing reasoning/asset views, add source/evidence claim links and test confidence/rationale quality.
7. Build safe next-collection recommendations with explicit expected information gain, coverage rationale and hard policy boundaries.
8. Improve recon orchestration: shared planner/worker context, stale/conflict handling, retry/cancellation/partial completion and duplicate prevention.
9. Build operator intelligence views for graph, coverage, evidence freshness, confidence and next actions.
10. Define and validate VEQ scoring, explanations and policy-aware recommendations against real normalized observations.

## Shared tool ecosystem and integrations

11. Validate runtime shared catalogue behavior for all agent types; refresh notifications and stale-state handling.
12. Verify agent capability profiles against configured live agents and real browser/MCP tool names.
13. Verify Yakit's actual installed version/interface, then implement and test a constrained adapter; add health, configuration, secret redaction, audit and cancellation.
14. Implement configured Kali environment discovery with executable path/version/package provenance; do not assume all Kali tools are installed.
15. Add evidence-based tool verification and measured performance collection (startup, run latency, reliability, failures, environment/version and freshness).
16. Complete external tool profile evidence, verified logos/artwork and readiness lifecycle without confusing catalogue visibility, operator approval and execution permission.

## Validation and release work (cross-cutting, not one job)

- Run gofmt, Go unit tests, full build, vet, frontend lint/typecheck/tests/build.
- Apply migrations on both clean and upgrade PostgreSQL databases; add migration tests for all recent migrations.
- Run DB-backed API and agent assembly tests, marketplace/collection/review smoke tests and authenticated authorization tests.
- Run end-to-end tests from real tool execution through raw evidence capture, parsing, normalized claims, asset/finding writes, graph/coverage/enrichment, API and UI.
- Review scope/target gates, interception, approval, egress, least privilege, secrets, audit, timeout, cancellation and retry behavior.
- Verify the regular ARTEX application still builds and runs with marketplace additions enabled and with external integrations absent/disabled.
- Update operator docs and distinguish committed, tested, verified, configured and operational.

## Current Job 5 implementation boundary

The new `enrichment` package is a deterministic, read-only projection over existing `db.Asset` records and the existing `assetgraph.Correlate` output. It surfaces recorded service names/types, HTTP metadata and technology labels with source-field rationale; conventional port-to-service hints are explicitly low-confidence and inferred. Relationships retain the graph's evidence, confidence and verification labels.

This is a foundation, not full integration: the projection is not yet wired into a DB-backed service, agent prompt/tool, API or operator UI; it does not create or persist new asset/finding records; and source-field labels are not immutable evidence claim IDs. Job 9 observation persistence and the validation backlog remain prerequisites for durable lineage.

## Counting rule

Count capability workstreams separately from implementation/validation sub-jobs. Keep partial items open until the remaining integration and verification criteria are met. The true remaining workload is therefore larger than the condensed countdown and should not be represented as a single precise number until the queue is decomposed into independently accepted tasks.
