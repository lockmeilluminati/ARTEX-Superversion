# ARTEX-Superversion platform architecture blueprint

This blueprint maps the README roadmap to package boundaries. It is a scaffold plan, not a claim
that every component is implemented or wired into the running server.

## Architectural rule

ARTEX remains the orchestration and authorization core. The existing planner, workers, task scope,
intercept/guard hooks, egress governor, traffic capture, evidence store and findings database remain
the authoritative execution and audit path. New packages provide metadata, adapters and intelligence;
they must not create a second execution engine or bypass existing controls.

## Package map

| Package | Responsibility | Current state |
|---|---|---|
| `integrations` | Adapter identity, health and future operation contracts | Initial contract + health probe |
| `integrations/yakit` | Yakit connector boundary | Health adapter; upstream interface pending verification |
| `integrations/kali` | Kali tool discovery boundary | Fixed allowlist version discovery |
| `toolregistry` | Versioned manifest and lifecycle registry | In-memory scaffold |
| `provenance` | Repository identity, ownership, release and lifecycle evidence | Data contracts + verifier interface |
| `discovery` | Pluggable candidate discovery and human review | Provider/review contracts |
| `tooldocs` | Source-backed generated documentation drafts | Profile contract + draft placeholder |
| `normalizer` | Adapter output to ARTEX-neutral findings/evidence | Initial mapping contract |
| `executionprofile` | Least-privilege execution profiles | Policy data model only |
| `veq` | Explainable recommendations and knowledge graph | Graph + recommender contracts |
| Existing `intercept`, `guard`, `egress` | Approval, target gating, and outbound controls | Reuse; no replacement |
| Existing `db`, `evidence`, `traffic` | Persistence, traffic and evidence | Reuse; future migrations/adapters required |
| Existing `server`, `web` | API and UI | Not yet wired to new packages |

## Data flow

1. Discovery providers return untrusted candidates with source references.
2. Provenance verification gathers independent signals and records evidence/confidence.
3. A human review decision moves a candidate through pending, accepted, rejected or quarantined.
4. An accepted manifest enters the registry; acceptance alone does not authorize installation or execution.
5. Adapter health reports connection state and supported capabilities.
6. A future operation request is checked against task scope, permissions, budgets and the existing approval gate.
7. Adapter results are normalized and linked to existing assets, findings, traffic and evidence records.
8. VEQ consumes verified registry metadata and ARTEX graph references to generate explainable recommendations.
9. Generated documentation remains a draft until reviewed and source-backed.

## Trust states

`discovered != verified != approved != installed != executable`.

Every transition must be explicit and auditable. Newly discovered tools are untrusted data.
Tool output and repository documentation must be treated as untrusted input. Never execute an
unverified candidate, package script, install hook or generated command.

## Integration boundary

Prefer documented local APIs, MCP, or explicit adapter protocols. Do not assume Yakit ports,
paths, authentication or schemas. For Kali, discovery is separate from execution. Later execution
must use fixed schemas, explicit target scope, approval, resource limits and suitable isolation.

## Next implementation slices

1. Verify Go scaffold compiles and add contract tests.
2. Verify Yakit's supported local automation interface and implement a real health probe.
3. Add server read-only status endpoint using existing auth/routing patterns.
4. Add UI integration status panel.
5. Implement one approved lab-only operation and normalized evidence ingestion.
6. Add registry persistence through a numbered DB migration.
7. Add provenance providers and discovery review workflow.
8. Connect VEQ recommendations to verified manifests and existing ARTEX graph data.
