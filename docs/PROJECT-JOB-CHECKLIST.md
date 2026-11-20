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

## 2. Tool registry and marketplace

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

## 3. Tool collections

- [x] Add database schema for named tool collections and collection membership.
- [x] Add database persistence functions to create, list, update, replace membership and delete collections. Commits: 11924acf136fc6fbe1e3ed2614fa01808510d86d and 2650ae0440c4b9aa53cc2e4a8f74a670e4ba1d77.
- [x] Add authenticated API endpoints for collection CRUD (GET/POST/PUT/DELETE; handlers use the authenticated PostgreSQL accessor).
- [x] Validate submitted membership against stable runtime/external catalogue IDs; reject unknown/blank IDs and de-duplicate entries. Existing saved IDs are retained if a catalogue entry is later removed; the UI should surface that stale state.
- [~] Add frontend collection management tab with create, rename/edit, description, add/remove catalogue tools and delete. Added stale-entry detection/removal; frontend lint/typecheck/build still need to run.
- [x] Make it explicit in the UI that collection membership is organizational only and grants no installation, enablement or execution permission.
- [~] Add API validation unit tests for accepted, duplicate, blank and unknown IDs. Database transaction behavior and authenticated HTTP CRUD integration tests remain pending. No test run has yet been confirmed.

## 4. External integrations and execution safety

- [~] Scaffold adapter contracts and initial Kali/Yakit integration profiles.
- [~] Document Yakit as the first external tool profile, with connectivity and supported operations explicitly unverified.
- [ ] Verify Yakit's installed-version interface and supported connection/authentication model before implementing calls. Do not assume ports, endpoints or API schemas.
- [ ] Implement and test a Yakit adapter only after interface verification.
- [ ] Wire integration adapters into server lifecycle, configuration and health reporting.
- [ ] Add operator-configurable connection settings with secret-safe storage and redacted API responses.
- [ ] Enforce least-privilege permissions, target/scope gates, approval requirements and audit/provenance records for external actions.
- [ ] Keep discovery/read-only operations separate from active scanning or other execution permissions.
- [ ] Add integration health, timeout, cancellation and failure handling.
- [ ] Add controlled integration tests using mocks first, then an explicitly configured local test instance.

## 5. VEQ intelligence and provenance

- [~] Scaffold VEQ, provenance, discovery, normalizer and execution-profile packages.
- [ ] Define the VEQ data model, scoring inputs, confidence representation and explanation format.
- [ ] Connect tool capabilities and execution profiles to VEQ recommendations.
- [ ] Normalize findings and assets from existing ARTEX tools and approved external adapters.
- [ ] Track provenance from source tool and execution through normalized findings and recommendations.
- [ ] Build explainable tool-selection and next-action recommendations.
- [ ] Add policy-aware workflow planning that respects scope, permissions and approval gates.
- [ ] Add end-to-end tests for recommendation quality, traceability and safety boundaries.

## 6. Quality, security and release readiness

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

1. Validate the current branch through CI: Go formatting/tests/build/vet and frontend lint/typecheck/build are configured; confirm a successful run and fix any failures.
2. Finish tool collections: build frontend management, stale-ID display and database/authenticated HTTP integration tests.
3. Validate the operator review workflow with migration tests and authenticated browser/API exercises; then add verified marketplace artwork using attributable upstream assets.
4. Verify Yakit's real interface, then implement and test its adapter.
5. Wire adapter lifecycle and permission gates into ARTEX.
6. Complete VEQ recommendations, normalization and provenance workflows.

## Status reporting rules

- "Committed" means the change exists in the repository.
- "Tested" means the relevant automated test or build command was actually run successfully.
- "Verified" means the relevant upstream/interface behavior was checked against reliable evidence or a real configured instance.
- "Operational" means configured and successfully exercised end-to-end in the target environment.

Do not mark a feature tested, verified or operational solely because its code has been committed.
