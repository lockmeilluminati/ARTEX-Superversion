# ARTEX-Superversion — Implementation Roadmap

This roadmap prioritizes ARTEX's core intelligence-gathering capability. The marketplace and external integrations are supporting infrastructure, not the product objective. Do not expand marketplace features while P0/P1 intelligence work remains incomplete.

## Priority 0 — Understand and reliably ingest what ARTEX already collects

1. **Collection pipeline inventory**
   - Trace planner/worker tool selection, tool execution, output parsing, asset/finding writes, evidence storage and frontend/API presentation.
   - Classify every path as wired, partial, scaffold-only or absent.
   - Identify the highest-value existing collection path and its gaps before adding more scanners.
   - Deliverable: a source-linked collection pipeline map and prioritized gap list.

2. **Canonical observation and evidence contract**
   - Build on the existing assets, findings, exploration graph and evidence stores; do not introduce a parallel graph.
   - Define normalized observations for domains, IPs, ports, services, URLs, technologies and findings.
   - Preserve source tool, run/task, timestamp, scope/target, confidence, raw-evidence reference and collection status.
   - Distinguish unknown, not attempted, failed, tested-negative and positively observed.
   - Keep raw evidence immutable or addressable; normalized records must retain a path back to it.

3. **Reliable, idempotent ingestion**
   - Implement/repair parsers for the most valuable existing tool outputs first.
   - Handle malformed, truncated, partial and empty outputs without losing valid observations.
   - Deduplicate repeated observations while preserving first-seen, last-seen and source history.
   - Add fixture-driven tests before broadening tool coverage.

## Priority 1 — Turn observations into useful intelligence

4. **Explainable asset correlation**
   - Correlate domains, IPs, ports, services, URLs, technologies and findings into the existing asset/exploration graph.
   - Preserve conflicting claims and their sources rather than overwriting them.
   - Mark inferred relationships as inferred, with confidence and supporting evidence.

5. **Collection coverage and freshness**
   - Track coverage by asset and discovery dimension.
   - Show what has been checked, what has not been attempted, what failed, what tested negative and what is stale.
   - Use coverage to avoid redundant collection and identify meaningful gaps.

6. **Evidence-backed enrichment**
   - Enrich asset context from observed service banners, HTTP metadata, DNS, TLS, technology fingerprints and other supported evidence.
   - Separate direct observations from derived conclusions.
   - Attach confidence rationale and provenance to every derived fact.

7. **Next-best collection recommendations**
   - Recommend safe, in-scope next steps based on asset context, coverage gaps, freshness and expected information gain.
   - Explain why a step is recommended and what evidence it is expected to produce.
   - Never bypass target/scope gates, approval requirements, least privilege or egress controls.

## Priority 2 — Orchestration and VEQ reasoning

8. **Recon orchestration**
   - Share collection context across planner and workers.
   - Avoid repeating completed work unless evidence is stale or contradictory.
   - Feed results back into the graph with traceable task/worker lineage.
   - Support cancellation, partial completion and retry without duplicate writes.

9. **VEQ intelligence**
   - Define scoring inputs, confidence representation and explanation format using actual normalized observations and provenance.
   - Recommend tools/actions based on capabilities, platform support, verification state and policy.
   - Make recommendations explainable and testable; do not treat a score as proof.

10. **Operator intelligence surfaces**
    - Present relationships, coverage gaps, evidence freshness, confidence and recommended next actions.
    - Make it easy to inspect why an asset or recommendation exists and which source observations support it.

## Supporting track — integrations and execution safety

- Existing ARTEX asset graph, exploration graph, findings store and planner/worker tools are the foundation to extend.
- Keep the adapter contracts and tool registry, but do not build more catalogue breadth for its own sake.
- Verify Yakit's actual installed-version interface, authentication and supported operations before implementing any adapter. Do not invent endpoints, ports or schemas.
- Any active external action must use ARTEX target/scope, approval, permission, egress and audit controls.
- Discovery/read-only permissions must remain distinct from active scanning or execution permissions.

## Maintenance track — marketplace and collections

The existing registry, catalogue, review workflow and tool collections remain available. Freeze new marketplace features, artwork and catalogue expansion until the P0/P1 intelligence milestones are delivered. Fix defects or security issues when necessary, and keep their validation status honest.

## Validation gates

- Parser and normalization fixtures cover valid, malformed, truncated, duplicate and partial results.
- Correlation tests preserve source lineage and conflicting observations.
- Coverage tests distinguish unattempted, failed, negative and positive observations.
- Recommendation tests prove scope/permission/approval constraints are respected.
- End-to-end tests trace a tool result through ingestion, persistence, graph correlation and operator presentation.
- Run Go/frontend checks and migration tests; record actual results rather than inferring success from committed code.

## Implementation status vocabulary

- **Scaffold**: types/interfaces exist; no production wiring.
- **Implemented**: code path exists and is covered by tests.
- **Integrated**: connected to ARTEX runtime and UI.
- **Verified**: tested against the actual external application/version.
- **Operational**: configured and successfully exercised end-to-end in the target environment.
