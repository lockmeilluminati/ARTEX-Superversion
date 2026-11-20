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
- [~] **P0 — Baseline the real collection pipeline.** `docs/COLLECTION-PIPELINE-BASELINE.md` now traces `insert_assets` through fail-closed asset interception, typed upserts, task provenance and auto-scope; documents Nmap/Nuclei/httpx/ffuf and BBot parser code; and maps the asset/finding APIs to their UI views. Both `ingest_bbot` and `ingest_scan_output` are included in the default worker tool set. Remaining: verify local scanner binary availability and command-to-file-to-ingestion operation, run parser/DB fixtures, and confirm asset-view refresh behavior.
- [~] **P0 — Define and integrate a canonical observation/evidence contract.** Added `normalizer.Observation`, explicit enums, provenance/evidence validation, stable identity/claim IDs and contract tests. Nmap/Nuclei/httpx/ffuf and BBot ingestion tools now return normalized observation projections alongside their existing writes; adapter tests cover provenance, port mapping and BBot scope separation. Still not durable observation persistence: run IDs are ARTEX ingestion-attempt IDs, evidence URIs are basename locator hints without immutable digest, and scanner scope remains unknown unless BBot explicitly establishes it. Run Go tests, review the live PostgreSQL/evidence schemas, and map projections to existing records before a migration. See `docs/OBSERVATION-EVIDENCE-CONTRACT.md`.
- [ ] **P0 — Make ingestion reliable and idempotent.** Add parsers/adapters for the highest-value existing discovery outputs; validate inputs, handle partial/malformed results, deduplicate repeated observations and update last-seen/source lineage without losing history.
- [ ] **P1 — Correlate the attack-surface graph.** Link domains↔IPs↔ports↔services↔URLs↔technologies↔findings with explainable edges and provenance. Resolve duplicates conservatively and retain conflicting observations rather than overwriting them.
- [ ] **P1 — Build collection coverage intelligence.** Track which assets and discovery dimensions have been checked, which are stale, what failed, and what remains unknown. Distinguish not-tested from tested-negative.
- [ ] **P1 — Add evidence-backed enrichment.** Derive service/technology identification and asset relationships from observations; label inferred relationships and include supporting evidence and confidence rationale.
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

The contract schema and identity rules are now committed. The immediate next job is to validate the contract against the actual scanner parsers and existing database model before introducing any persistence migration.

1. **P0 — Validate and integrate the canonical observation contract.** Run `go test ./normalizer`; add adapter mapping from Nmap/httpx/Nuclei/BBot parser results into `normalizer.Observation`; verify stable IDs, evidence references, scope and collection states with fixtures. Review the proposed append-only claim-link migration against current asset/evidence/finding schemas before applying it.
2. **P0 — Verify the collection pipeline operationally.** Confirm local scanner binary availability and command → output file → worker parser → persistence, run parser/DB fixtures, and validate asset UI refresh as detailed in `docs/COLLECTION-PIPELINE-BASELINE.md`.
3. **P0 — Make ingestion reliable and idempotent against the agreed contract.** Add/adjust adapters, partial/malformed handling, deduplication and source lineage without losing history.
4. **P1 — Correlate the attack-surface graph.** Add explainable links with provenance and retain contradictory observations.
5. **P1 — Build collection coverage intelligence.** Distinguish not-tested, tested-negative, partial and failed collection.
6. **P1 — Add evidence-backed enrichment and safe next-collection recommendations.**
7. **P0 — Validate runtime capability profiles against live configured agents.** Profile persistence, authenticated API, operator UI and explicit planner compatibility rules are committed; live-instance validation and DB-backed tests remain.
8. **P0 — Make Yakit the first live external integration.** Verify the actual installed version and interface before implementing supported adapter operations.
9. **P0 — Implement Kali OS discovery.** Discover installed tools and collect version/path/package provenance; do not infer installation from catalogue metadata.
10. **P2 — Improve recon orchestration, operator intelligence views and VEQ.**
11. **Maintenance:** validate marketplace, collections and review workflow; continue only necessary defects/security fixes.

## Status reporting rules

- "Committed" means the change exists in the repository.
- "Tested" means the relevant automated test or build command was actually run successfully.
- "Verified" means the relevant upstream/interface behavior was checked against reliable evidence or a real configured instance.
- "Operational" means configured and successfully exercised end-to-end in the target environment.

Do not mark a feature tested, verified or operational solely because its code has been committed.
