# ARTEX-Superversion — Project Job Checklist

Use this file as the living project punch list. Tick an item only when its code/docs are committed and the stated acceptance criteria are met. A committed change is **not** automatically tested or deployed.

Legend:
- [x] Implemented and committed (tests/build may still be pending)
- [~] Partially implemented; remaining work is listed
- [ ] Not started

Last reviewed: 2026-11-20

## 1. Project foundation

- [x] Establish project vision, roadmap, architecture and integration documentation.
- [x] Preserve upstream ARTEX attribution and document the relationship to the upstream project.
- [x] Add BackTrack historical logo to the README with attribution.
- [x] Scaffold foundational packages for integrations, tool registry, provenance, discovery, documentation, normalization, VEQ and execution profiles.
- [x] Document the integration plan, platform architecture, implementation roadmap, test matrix and registry migration approach.

## 2. Core intelligence gathering (highest priority)

The marketplace is a core operator and agent-facing surface for a shared tool ecosystem. ARTEX must discover, verify, describe and safely expose tools to every suitable agent. New development must also improve intelligence gathering, evidence correlation and prioritization. Build on the existing ARTEX asset graph, exploration graph, findings store and worker/planner tools; do not create a parallel knowledge store.

- [x] Reuse the existing persistent asset model for domains, IPs, ports, services, URLs, technologies and related observations (`db/assets.go`).
- [x] Reuse the existing exploration graph, planner/worker lineage and findings persistence (`agent/tools.go`, `db/findings.go`). This is existing ARTEX foundation, not a newly completed Superversion feature.
- [~] **P0 — Baseline the real collection pipeline.** `docs/COLLECTION-PIPELINE-BASELINE.md` traces `insert_assets` through fail-closed interception, typed upserts, task provenance and auto-scope; documents the scanner parsers and asset/finding UI paths; and records the worker ingestion tools. Added `scripts/inventory-collection-tools.sh` plus a repeatable operator procedure for binary/version inventory, parser/DB tests, authorized-lab command-to-file-to-ingestion checks, replay/scope checks and UI refresh. Still not operationally verified: run the inventory on the actual worker host, execute parser/PostgreSQL fixtures, exercise each installed scanner through ingestion, and confirm asset-view refresh.
- [~] **P0 — Define and integrate a canonical observation/evidence contract.** Added `normalizer.Observation`, explicit enums, provenance/evidence validation, stable identity/claim IDs and scanner adapter projections. Migrations 0016–0017 and `db/observation_provenance.go` now provide collection-run, evidence-metadata and append-only claim persistence primitives; collection runs have an explicit one-way `planned → running → terminal` lifecycle (finalization is allowed only from `running` and only once; terminal records cannot be overwritten), identical claim replay is a no-op and same-ID payload changes conflict. PostgreSQL tests cover lifecycle transitions, replay and contradictory claims but have not been run. Added `evidence.Store.StoreCollectionArtifact` to import run-relative regular files into the existing content-addressed evidence store, compute SHA-256/length during copy, register opaque metadata only when task/tool identity matches the collection run, and retain registered blobs during GC; focused path/size tests are committed but unrun. Still pending: wire real process launch/exit IDs through worker context, replace adapter attempt IDs/locator hints with registered evidence IDs/digests, persist normalized claims during ingestion, add transactional links to existing records, and execute full FK/rollback/scope/lifecycle tests. Source inspection of pinned Norma v0.4.11 confirms three distinct launch paths (managed Bash, synchronous Bash fallback and interactive PTY); tool-level Pre/Post hooks cannot identify individual process lifecycles. The exact required observer seam and ARTEX wiring sequence are documented in `docs/NORMA-PROCESS-LIFECYCLE-INTEGRATION.md`. Do not treat scanner observations as durably recorded until that integration is complete. See `docs/OBSERVATION-PERSISTENCE-DESIGN.md` and `docs/OBSERVATION-EVIDENCE-CONTRACT.md`.
  - [x] Claim insertion now verifies that the registered evidence artifact, collection run, task ID and source-tool ID all agree; PostgreSQL regression cases reject mismatched task/tool claims (commits 80309d9 and 26b19ab; tests not executed).
- [~] **P0 — Make ingestion reliable and idempotent.** Existing Nmap/Nuclei/httpx/ffuf and BBot parsers validate typed fields and bound line/file/record sizes; BBot's 50000-record cap reports excess records as `over_cap` with `truncated=true` (commits e096638, 9d7cddc, 7407082, f5c71b1, 1c8b763, a65e1ae, bc4870c). Added `ExplorationStore.AddFactIfAbsent`: transaction-scoped PostgreSQL advisory locking deduplicates by canonical source+summary+detail across adapters, returns one fact ID for concurrent/replayed writes, and unions replay anchors; different details or sources remain distinct (commit 7b45917, refined in bfe9212). `record_fact` uses this persistence boundary (a63a42d). Scanner and BBot ingestion no longer rely on a read-then-write in-memory/database scan; every candidate goes through the atomic check, preserving anchor union on replay and reporting inserted versus existing facts (98f47db). `recordOneFact` now returns its own inserted/replayed result so per-ingestion counts do not infer status from a shared mutable counter (38c485d, 43460f1); removed the obsolete scan-before-write path. Added a 12-writer concurrency test, metadata-insensitive canonical replay assertion and distinct-detail assertion (901531b, extended in dff30b5; not run). Still pending: preserve all differing evidence/source-run lineage when a canonical fact is replayed, update last-seen metadata, consistent partial-result reporting, and execution of parser/PostgreSQL integration tests.
- [~] **P1 — Correlate the attack-surface graph.** Added a deterministic, read-only `assetgraph.Correlate` projection for explicit domain/root-domain, subdomain/IP, service/host and endpoint/service relationships. Every edge carries a relation, source-field rationale, confidence and observed/inferred label; company membership and subnet proximity alone never create edges. Unit fixtures cover supported links, conservative non-links and deterministic output. Remaining: connect persisted evidence/observation claims and findings, expose graph through authenticated API/UI, and run tests against real DB fixtures.
- [~] **P1 — Build collection coverage intelligence.** Reused ARTEX's existing declared-scope asset denominator, untested-asset backlog, worker `record_coverage` tool and planner/wrap-up ledger. Added an explicit `test_failed` outcome (always an open gap), deterministic state assessment (`not_tested`, `tested_positive`, evidence-backed `tested_negative`, `not_applicable`, `partial`, `failed`, `stale`, `unknown`), and a conservative 30-day read-time freshness default while retaining the previous state for stale entries. Added migration 0015 and clean-schema support; see `docs/COLLECTION-COVERAGE-INTELLIGENCE.md`. Remaining: expose assessments in the operator UI, improve dimension-specific asset coverage beyond the current coarse fact-anchor signal, connect evidence to durable observation claims (Job 9), and run Go/PostgreSQL tests and migrations.
- [~] **P1 — Add evidence-backed enrichment.** Added a deterministic read-only `enrichment.Derive` projection over existing asset records and `assetgraph.Correlate`: exposes recorded service names/types, HTTP metadata and technologies with source-field rationale; conventional port hints are explicitly low-confidence/inferred; relationship projections retain evidence/confidence/verification. Remaining: link outputs to immutable observation/evidence claims, integrate with existing reasoning/API/operator UI, and test confidence/rationale against real evidence. See `docs/WORK-QUEUE-AUDIT.md`.
- [ ] **P1 — Generate useful next-collection actions.** Recommend the next safe, in-scope collection step based on coverage gaps, asset context, expected information gain and tool capability. Explain why; never bypass target/scope/approval/egress controls.
- [ ] **P2 — Improve recon orchestration.** Have planner/worker runs share collection context, avoid repeating completed work, detect stale or contradictory observations, and feed results back into the graph with traceable lineage.
- [ ] **P2 — Add operator-facing intelligence views.** Surface asset relationships, collection coverage, evidence freshness, confidence and recommended next steps in the existing UI/API.
- [ ] Add fixture-based parser tests, normalization/deduplication tests, graph-correlation tests, coverage tests and end-to-end tests from tool output to persisted asset/evidence.

## 3. Tool registry and marketplace

- [x] Add the tool registry database migration and seed the initial Yakit profile.
- [x] Add database functions to upsert and list registry entries.
- [x] Define the tool manifest contract and conservative Yakit metadata profile.
- [x] Add the unified read-only marketplace API at GET /api/tool-registry.
- [x] Ensure marketplace projection does not expose custom execution specifications or secrets.
- [x] Connect the frontend API client and TypeScript registry types.
- [x] Add the External Ecosystem marketplace tab.
- [x] Add external tool profile sheets with lifecycle, availability, trust, license, maintainer, specialties, permissions and source details.
- [x] Add category filtering with categories derived dynamically from registered external entries. Commit: c52d1c8c9adbe3e42f935ec69bbf960cbeb36ba9.
- [~] Add individual tool artwork. Current cards use generic CSS artwork; add only verified tool logos/images and a safe fallback.
- [x] Add marketplace sorting and side-by-side comparison (name, verification, trust, lifecycle, license; capabilities, platforms, license, trust, readiness and availability). Commits: a46f8919c7e355edd6e93c8ae3b86b2873cb38da and bd4ab5e4a56c5059acc6693de931519b83ce2b08.
- [~] Expand the external catalogue with evidence-backed candidate profiles for Nmap, Nuclei, httpx, Katana, ffuf and Dalfox (migrations 0011 and 0012; docs/EXTERNAL-TOOL-CATALOG.md). Upstream sources, licenses and capabilities are recorded; all six remain disabled, unverified candidates pending local installation/interface checks and adapter tests.
- [x] Add an authenticated operator review workflow for external registry metadata (approve/reject/reopen, required rejection reason, latest state and audit history). Approval is metadata-only and never enables, installs, verifies or grants execution. Migration 0013; see docs/TOOL-REVIEW-WORKFLOW.md. API validation tests added; DB migration/integration and frontend build validation pending.
- [~] **P0 — Establish one shared tool registry for all agents.** The read-only `shared_tool_catalogue` agent tool now exposes live metadata for native/custom runtime tools and external registry profiles to every augmented agent. It omits execution specs/secrets and explicitly does not grant execution. A search-matching unit test is committed; runtime integration tests and discovery refresh notifications remain. Per-agent callable-tool bindings remain separate. See docs/AGENT-TOOL-CAPABILITIES.md.
- [x] **P0 — Define mandatory tool readiness/verification metadata.** Manifest now supports installation, interface verification, verification timestamp/version, execution mode (CLI, GUI, API, MCP, hybrid), required agent capabilities and measured performance. Missing values remain unknown; listed, installed, verified, approved and operational are distinct. See docs/AGENT-TOOL-CAPABILITIES.md.
- [x] **P0 — Show readiness state on all tool cards and external profiles.** External cards show declared execution mode, verification and installation status; native/custom cards explicitly show that mode and verification are not yet classified. Profiles show agent requirements and measured performance when available. Unknown is shown rather than guessed.
- [~] **P1 — Add evidence-based execution/performance profiles.** Manifest and UI support sample count and median startup/run measurements; measurement collection, freshness and failure history are not yet wired.
- [~] **P1 — Make agent-tool compatibility machine-readable.** Added AgentProfile, capability vocabulary and deterministic MatchAgent logic. Runtime assembly persists evidence-based terminal/platform and connected-MCP/browser profiles; `shared_tool_catalogue` returns per-agent compatibility results. Authenticated `GET /api/agent-capabilities` and the System → Tools → Agent capabilities view now expose timestamps/evidence. Remaining: run DB-backed assembly/API tests, validate detected names against the configured browser MCP version and confirm planner behavior against live agent prompts.

## 4. Tool collections

- [x] Add database schema for named tool collections and collection membership.
- [x] Add database persistence functions to create, list, update, replace membership and delete collections. Commits: 11924acf136fc6fbe1e3ed2614fa01808510d86d and 2650ae0440c4b9aa53cc2e4a8f74a670e4ba1d77.
- [x] Add authenticated API endpoints for collection CRUD (GET/POST/PUT/DELETE; handlers use the authenticated PostgreSQL accessor).
- [x] Validate submitted membership against stable runtime/external catalogue IDs; reject unknown/blank IDs and de-duplicate entries. Existing saved IDs are retained if a catalogue entry is later removed; the UI should surface that stale state.
- [~] Add frontend collection management tab with create, rename/edit, description, add/remove catalogue tools and delete. Added stale-entry detection/removal; frontend lint/typecheck/build still need to run.
- [x] Make it explicit in the UI that collection membership is organizational only and grants no installation, enablement or execution permission.
- [~] Add API validation unit tests for accepted, duplicate, blank and unknown IDs. Database transaction behavior and authenticated HTTP CRUD integration tests remain pending. No test run has yet been confirmed.

## 5. External integrations and execution safety

- [~] Scaffold adapter contracts and initial Kali/Yakit integration profiles.
- [~] Document Yakit as the first external tool profile, with connectivity and supported operations explicitly unverified.
- [ ] Verify Yakit's installed-version interface and supported connection/authentication model before implementing calls. Do not assume ports, endpoints or API schemas.
- [ ] Implement and test a Yakit adapter only after interface verification.
- [ ] Wire integration adapters into server lifecycle, configuration and health reporting.
- [ ] **P1 — Add shared tool availability/change notifications.** When tools are discovered, upgraded, removed or re-verified, refresh the shared registry so all agents see current state without stale private catalogues.
- [ ] Add operator-configurable connection settings with secret-safe storage and redacted API responses.
- [ ] Enforce least-privilege permissions, target/scope gates, approval requirements and audit/provenance records for external actions.
- [ ] Keep discovery/read-only operations separate from active scanning or other execution permissions.
- [ ] Add integration health, timeout, cancellation and failure handling.
- [ ] Add controlled integration tests using mocks first, then an explicitly configured local test instance.

## 6. VEQ intelligence and provenance

- [~] Scaffold VEQ, provenance, discovery, normalizer and execution-profile packages.
- [ ] Define the VEQ data model, scoring inputs, confidence representation and explanation format.
- [ ] Connect tool capabilities and execution profiles to VEQ recommendations.
- [ ] Normalize findings and assets from existing ARTEX tools and approved external adapters.
- [ ] Track provenance from source tool and execution through normalized findings and recommendations.
- [ ] Build explainable tool-selection and next-action recommendations.
- [ ] Add policy-aware workflow planning that respects scope, permissions and approval gates.
- [ ] Add end-to-end tests for recommendation quality, traceability and safety boundaries.

## 7. Quality, security and release readiness

- [~] Add initial unit/security tests for registry manifest and API projection behavior, including operator review request validation. Database migration/integration tests remain pending.
- [ ] Run gofmt on changed Go files.
- [ ] Run go test ./....
- [ ] Run go build ./....
- [ ] Run go vet ./....
- [ ] Run frontend lint and production build from web/.
- [ ] Apply database migrations against a clean test database and an upgrade database.
- [~] CI workflow already runs Go formatting/build/vet/tests, frontend typecheck/tests/static build and SQL index-script checks; frontend lint was added in commit 91a59ad223db938b35b289163f729199c4932dc3. Confirm a successful run and add dedicated migration upgrade checks.
- [ ] Review API authentication/authorization for all new registry and collection routes.
- [ ] Perform a final security review of external-tool configuration, execution gates, secrets and audit trails.
- [ ] Update README and operator documentation to distinguish implemented, tested, configured and verified capabilities.

## Current next-up sequence

The canonical observation schema and first parser projections are committed. The next job is operational verification; no local tests or GitHub Actions runs have yet been confirmed for this checkpoint.

### Readiness checkpoint (2026-11-20)

- GitHub Actions: the repository Actions API currently reports **zero workflow runs**. The CI workflow is defined, but there is no passing CI result to rely on.
- Local Go tests/build: not run. A source checkout could not be obtained in this execution environment because GitHub DNS/network access was unavailable.
- Marketplace: the tool-registry API, marketplace UI, collections and review controls are present in source. They remain **implemented but unverified**; external tool profiles are metadata-only and do not install or execute tools.
- Regular application: the existing application source and CI build steps remain present, but a fresh build has not been demonstrated in this checkpoint.
- Observation adapters: return normalized projections alongside existing writes; they do not yet persist canonical observation claims. The current run/evidence identifiers are explicitly provisional.

### Current next-up sequence

1. **P0 — Verify the collection pipeline operationally (NEXT).** Run `go test ./normalizer ./agent`, then the full Go build/test/vet suite; validate scanner binaries and versions; exercise command → output → parser → adapter → existing DB writes; verify asset UI refresh.
2. **P0 — Complete observation provenance and persistence design.** Replace ingestion-attempt IDs with actual execution IDs where available; replace basename evidence hints with immutable evidence-store references/digests; review the proposed claim-link mapping against existing DB schemas. Do not apply a migration until evidence storage and lineage are verified.
3. **P0 — Make ingestion reliable and idempotent.** [~] Shared exact-payload replay deduplication now covers scanner facts and BBot third-party intelligence facts; a PostgreSQL replay test is added. Remaining: DB-level concurrency-safe idempotency, conflict/last-seen/source-lineage persistence, and full malformed/partial/truncated/replay test execution.
4. **P1 — Correlate the attack-surface graph.** Add explainable links with provenance and retain contradictory observations.
5. **P1 — Build collection coverage intelligence.** Distinguish not-tested, tested-negative, partial and failed collection.
6. **P1 — Add evidence-backed enrichment and safe next-collection recommendations.**
7. **P0 — Validate runtime capability profiles against live configured agents.** Profiles, authenticated API and operator UI are committed; live-instance validation and DB-backed tests remain.
8. **P0 — Make Yakit the first live external integration.** Verify the actual installed version and interface before implementing supported adapter operations.
9. **P0 — Implement Kali OS discovery.** Discover installed tools and collect version/path/package provenance; do not infer installation from catalogue metadata.
10. **P2 — Improve recon orchestration, operator intelligence views and VEQ.**
11. **Maintenance — Validate marketplace, collections and review workflow.** Fix defects/security issues; no further feature expansion until intelligence milestones progress.

This is a **condensed sequence of workstreams, not a complete job countdown**. It is not an accurate count of all remaining work: the roadmap contains 16 broad capability workstreams plus cross-cutting testing, security, integration and release tasks, and several numbered entries bundle multiple independently deliverable jobs. See `docs/WORK-QUEUE-AUDIT.md` for the expanded remaining backlog.

## Status reporting rules

- "Committed" means the change exists in the repository.
- "Tested" means the relevant automated test or build command was actually run successfully.
- "Verified" means the relevant upstream/interface behavior was checked against reliable evidence or a real configured instance.
- "Operational" means configured and successfully exercised end-to-end in the target environment.

Do not mark a feature tested, verified or operational solely because its code has been committed.


### Job 10 execution attempt — 2026-11-20

- Re-read the current `.github/workflows/ci.yml`: the workflow is configured for pushes to `main` and pull requests. It includes the Go/PostgreSQL, shell, frontend and Python jobs described above.
- Re-queried the repository Actions API: **0 workflow runs**. There is no run result or test log to inspect, and the available GitHub connection exposes read-only workflow inspection/retry actions but no workflow-dispatch action.
- Checked the execution environment: Go is `1.23.2`, while `go.mod` requires Go `1.26.10`; Node is `22.16.0` and npm is `10.9.2`. The repository could not be cloned because this environment cannot resolve `github.com`.
- Re-inspected `agent/bbot_ingest_tool.go` and `agent/scan_ingest_tool.go`; the required imports are present in the current branch. This is a source inspection only, not a compile or test result.
- **Job 10 remains blocked / not passed.** No Go tests, build, vet, frontend lint/typecheck/build, migration test, or marketplace runtime test was executed. Do not count this job down until a CI run or a suitable local environment provides successful evidence.


### Job 9 progress — provenance and persistence design (2026-11-20)

- [x] Review the current normalizer contract, scanner adapters, `RunInfo`, evidence store, asset schema and existing exploration persistence boundaries.
- [x] Specify the collection-run lifecycle, with IDs created at actual process launch and propagated into ingestion; ingestion-attempt IDs must not masquerade as process IDs.
- [x] Specify generic immutable scanner-artifact registration using the existing content-addressed evidence blob foundation, including digest/length verification, path confinement, retention and secret-safety requirements.
- [x] Specify append-only observation claims linked to existing assets/findings/exploration nodes; identity keys are non-unique so contradictory claims survive.
- [x] Specify replay/conflict semantics, transaction expectations, migration sequencing and acceptance tests.
- [ ] Implement and test the collection-run lifecycle and actual run-ID propagation.
- [ ] Implement generic evidence artifact metadata/registration and scanner artifact integration.
- [ ] Review exact foreign-key targets and task/exploration lifecycle against the full schema before writing migrations.
- [ ] Implement append-only observation claim persistence and transaction-aware projection links.
- [ ] Add PostgreSQL integration tests for replay, conflicts, rollback and projection deletion.
- [ ] Apply and verify migration only after schema review and tests.

Design document: `docs/OBSERVATION-PERSISTENCE-DESIGN.md`. No migration was created or applied. **Job 9 remains in progress; design is committed, implementation and DB verification are not complete.**


### Job 8 progress — ingestion reliability and idempotency (2026-11-20)

- [x] Consolidate scanner and BBot fact persistence on one exact-replay deduplication path.
- [x] Apply the same deduplication rule to BBot third-party intelligence facts; repeated ingestion of the same source, summary and detail is counted as existing rather than appended again.
- [x] Preserve distinct facts when their source, summary or detail differs; this avoids silently overwriting differing observations.
- [x] Add PostgreSQL-backed BBot assertions for exact replay (zero new facts) and distinct payload preservation (same summary, different detail remains a separate fact).
- [~] Existing asset upserts remain idempotent through the asset store; scanner and BBot parsers already bound file/line sizes and preserve complete records from supported partial/truncated inputs.
- [ ] Add database-enforced, concurrency-safe idempotency (the current shared helper checks the exploration graph before writing and is not an atomic uniqueness guarantee under concurrent ingesters).
- [ ] Verify last-seen and source-lineage update semantics without erasing contradictory source claims; complete observation-claim persistence remains gated by Job 9.
- [ ] Execute parser and PostgreSQL integration tests, including malformed, oversized, truncated, replayed and conflicting fixtures. No tests were run for this commit.
- [ ] Add end-to-end partial-failure and cancellation/retry tests, then reconcile result counters with actual persisted outcomes.

**Job 8 is in progress, not complete.** The change closes the BBot replay gap, but the graph-level check-then-insert is not concurrency-safe and the provenance/claim layer is not yet durable.


### Job 7 progress — explainable attack-surface correlation (2026-11-20)

- [x] Inspect the existing asset schema and confirm implicit domain/IP/service relationships already present in asset fields; do not create a second asset inventory.
- [x] Add a pure deterministic correlation projection over existing `db.Asset` records.
- [x] Add explainable edge metadata: source/target asset IDs, relation, exact supporting field rationale, confidence and observed/inferred status.
- [x] Correlate root-domain membership, subdomain A/AAAA/IP links, service/endpoint host links and endpoint-to-HTTP-service origin links.
- [x] Keep conservative boundaries: no edge from shared company, task, subnet proximity or technology label alone; no asset records are mutated.
- [x] Add unit fixtures for expected links, no unsupported company/subnet links and deterministic ordering.
- [ ] Add evidence/observation-claim IDs to edges once Job 9 persistence exists; retain multiple contradictory claims as separate provenance paths.
- [ ] Add finding-to-asset links using the existing findings asset references; do not infer vulnerability-to-host links from summary text.
- [ ] Add bounded database loading and an authenticated graph API with pagination/limits.
- [ ] Add operator graph view and filters for observed vs inferred, confidence and evidence.
- [ ] Run unit and PostgreSQL integration tests; no tests were executed for this commit.

**Job 7 is in progress.** The initial correlation engine is committed, but it is a source-field projection only: edges are not persisted, evidence claim IDs are not yet attached, and no API/UI exposes the graph.


### Job 6 progress — collection coverage intelligence (2026-11-20)

- [x] Reuse the existing ARTEX asset-coverage denominator/backlog and risk-class ledger; do not create a second asset or coverage store.
- [x] Preserve the distinction between a missing ledger entry (not tested) and an evidence-backed negative result.
- [x] Add explicit `test_failed` ledger outcome and classify it as a gap, never as a negative result.
- [x] Add pure assessment states for not tested, positive, evidence-backed negative, not applicable, partial, failed, stale and unknown; stale assessments retain their previous state.
- [x] Add a 30-day default read-time freshness assessment without rewriting historical timestamps; callers of the pure helper can choose another window.
- [x] Add clean-schema support and additive migration 0015; document compatibility with the normal ARTEX task/agent execution path.
- [ ] Add authenticated operator API/UI presentation and filters for state/freshness.
- [ ] Replace or supplement the coarse fact-anchor asset-tested signal with dimension-specific evidence-backed coverage; a fact anchor alone does not prove each discovery dimension was tested.
- [ ] Link coverage evidence to durable observation/evidence claim IDs after Job 9 persistence.
- [ ] Run Go unit tests, PostgreSQL integration tests and clean/upgrade migration checks. None were executed for this work.

**Job 6 is in progress, not complete.** This is an additive backend improvement to the existing ARTEX coverage system; the normal program and marketplace execution/permissions are unchanged.


### Job 5 progress — evidence-backed enrichment (2026-11-20)

- [x] Add a deterministic read-only projection over existing ARTEX asset records.
- [x] Surface explicitly recorded service names/types, HTTP status/title and technology labels with source-field rationale.
- [x] Add conservative conventional-port service hints, always labeled inferred and low confidence; a port number alone is not treated as service identification.
- [x] Reuse existing explainable asset graph relationships and retain evidence/rationale, confidence and observed/inferred verification labels.
- [ ] Link projections to immutable evidence/observation claim IDs once Job 9 persistence is implemented.
- [ ] Integrate enrichment into the existing agent reasoning, authenticated API and operator intelligence UI.
- [ ] Add evidence/confidence tests against real DB fixtures and end-to-end tool-output-to-enrichment tests; run Go tests/build.
- [ ] Verify no changes regress the regular ARTEX execution path.

**Job 5 is partially implemented, not complete.** The new `enrichment` package is a read-only projection and does not persist or promote inferred values into canonical assets/findings. Full remaining-work audit: `docs/WORK-QUEUE-AUDIT.md`.
