# Roadmap implementation map

This file translates the public roadmap into buildable work packages.

## Phase 1 — Foundation & trust
- Baseline: preserve existing ARTEX tests, build and CI.
- Integration contracts: `integrations/`.
- Trust boundaries: reuse `intercept/`, `guard/`, `egress/`, `evidence/`.
- Acceptance: clean build, no behavior change when new integrations are disabled.

## Phase 2 — Unified tool registry
- Manifest schema and lifecycle: `toolregistry/`.
- Provenance record and verification: `provenance/`.
- Persistence: future numbered migration in `db/migrations/`; do not edit published migrations.
- Search/filter API and UI are future work.

## Phase 3 — Yakit
- Verify local supported interface on a real installation.
- Add health and capability negotiation to `integrations/yakit/`.
- Add supported result/traffic mapping to `normalizer/`, `evidence/`, and `traffic/`.
- Never infer API endpoints or claim compatibility without tests.

## Phase 4 — Kali
- Expand package metadata using authoritative Kali/Debian package sources.
- Link package names to canonical upstream repositories with confidence/evidence.
- Execution profiles live in `executionprofile/`; actual execution must reuse ARTEX approval and target controls.
- Discovery does not imply installation or permission.

## Phase 5 — Discovery and documentation
- Provider interfaces: `discovery/`.
- Provenance and lifecycle signals: `provenance/`.
- Draft profiles and source references: `tooldocs/`.
- Review queue before trust, installation or execution.

## Phase 6 — VEQ
- Knowledge graph contracts: `veq/`.
- Recommendations must explain capability, platform, verification and policy reasons.
- Connect to ARTEX asset/exploration graphs rather than duplicating them.
- Future dashboards in `web/` and read-only APIs in `server/`.

## Phase 7 — Hardening
- Unit, contract, integration and end-to-end tests.
- Reproducible builds, SBOM and release metadata.
- Compatibility/support matrix and security review.
- Higher autonomy remains disabled until scope, approval, isolation and audit paths are verified.

## Status vocabulary
- **Scaffold**: types/interfaces exist; no production wiring.
- **Implemented**: code path exists and is covered by tests.
- **Integrated**: connected to ARTEX runtime and UI.
- **Verified**: tested against the actual external application/version.
