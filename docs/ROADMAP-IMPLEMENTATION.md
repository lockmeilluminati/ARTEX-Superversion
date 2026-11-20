# ARTEX-Superversion — Implementation Roadmap

ARTEX-Superversion is a shared intelligence and tool-orchestration platform. Its agents share one live tool catalogue, understand how tools are operated, and select compatible tools/agents based on verified capabilities. Yakit is the first live external integration; Kali OS is the first broad local tool-discovery source. Marketplace cards expose truthful readiness and verification state.

## Priority 0 — Shared tool ecosystem and first integrations

1. **One shared tool registry for all agents**
   - Make the registry the single source of truth for native, custom, external and discovered Kali tools.
   - Ensure planner, workers and future agents read the same current catalogue and capability metadata.
   - Newly discovered or updated tools become visible to all agents without manual per-agent registration.
   - Keep visibility separate from permission to execute.

2. **Mandatory verification and agent-readiness contract**
   - Track installation state, detected version, executable/interface, verification state and timestamp, platform, dependencies, upstream/source provenance and license where known.
   - Classify execution mode as CLI, GUI, API, MCP or hybrid.
   - Declare required agent capabilities: terminal/command execution, screenshots and visual interaction, browser control, API/MCP connectivity, credentials and OS/dependencies.
   - Distinguish listed, installed, interface-verified, approved, available and operational. Unknown must remain unknown; never infer verified from catalogue metadata.
   - Display these properties on every marketplace card and full profile, with clear safe fallbacks for missing artwork.

3. **Yakit — first live external integration**
   - Detect local Yakit installation/version where configured.
   - Verify the actual installed build's documented interface, authentication and supported operations before writing adapter calls. Do not invent ports, endpoints or schemas.
   - Add a health/connectivity check and controlled capability verification.
   - Implement the adapter only against verified interfaces; publish its capabilities and GUI/API/CLI requirements to the shared registry.
   - Preserve target/scope, approval, least-privilege, egress and audit controls.

4. **Kali OS tool discovery**
   - Discover tools actually installed on a configured Kali environment; record executable paths, versions and package provenance where available.
   - Do not assume every Kali package is installed or that package presence proves an interface works.
   - Populate/update shared registry entries idempotently with source, license (when verified), capabilities, platform, execution mode, dependencies and status.
   - Add controlled per-tool verification and distinguish missing, unsupported, failed, installed and verified states.
   - Keep discovery read-only; installation and execution require separate explicit operator policy.

5. **Baseline the real collection pipeline**
   - Trace planner/worker selection, execution, parsing, asset/finding writes, evidence storage and UI/API presentation.
   - Classify paths as wired, partial, scaffold-only or absent; identify duplicated paths and ingestion gaps.
   - Deliver a source-linked pipeline map before adding more scanners.

6. **Canonical observation and evidence contract**
   - Extend existing assets, findings, exploration graph and evidence stores; do not create a parallel knowledge store.
   - Define normalized observations for domains, IPs, ports, services, URLs, technologies and findings.
   - Preserve source tool, source run/task/agent, timestamp, scope/target, confidence, raw-evidence reference and collection status.
   - Distinguish unknown, not attempted, failed, tested-negative and positively observed.

7. **Reliable, idempotent ingestion**
   - Implement/repair parsers for highest-value existing tool outputs first.
   - Handle malformed, truncated, partial and empty outputs without losing valid observations.
   - Deduplicate repeated observations while preserving first-seen, last-seen and source history.
   - Add fixture-driven tests before broadening tool coverage.

## Priority 1 — Agent compatibility and useful intelligence

8. **Agent capability declarations and tool matching**
   - Agents declare supported execution modes and requirements, such as CLI, GUI/screenshot, browser, API/MCP, OS and dependencies.
   - Planner selects a compatible agent/tool pairing and explains incompatibility or setup requirements.
   - A GUI tool must not be assigned to an agent lacking required screen/visual interaction capabilities.
   - A CLI tool must not be assumed fast solely because it is CLI-based.

9. **Measured execution/performance profiles**
   - Record observed startup and operation latency, reliability, timeout/failure history and environment/version.
   - Classify performance from measurements with sample count and freshness, not guesses.
   - Keep performance claims distinct from interface verification and security approval.

10. **Shared catalogue freshness**
    - Refresh agent-visible tool state when tools are discovered, installed, upgraded, removed or re-verified.
    - Prevent stale per-agent copies; retain audit history of registry changes.

11. **Explainable asset correlation**
    - Correlate domains, IPs, ports, services, URLs, technologies and findings into the existing graph.
    - Preserve conflicting claims and their sources rather than overwriting them.
    - Mark inferred relationships as inferred, with confidence and supporting evidence.

12. **Collection coverage and freshness**
    - Track coverage by asset and discovery dimension.
    - Distinguish unattempted, failed, tested-negative, positive and stale.
    - Use coverage to avoid redundant collection and identify meaningful gaps.

13. **Evidence-backed enrichment and next-best collection**
    - Enrich context from observed service banners, HTTP metadata, DNS, TLS and technology fingerprints.
    - Recommend safe, in-scope next steps based on context, coverage, freshness and expected information gain.
    - Explain expected evidence and never bypass scope, permission, approval, least-privilege or egress controls.

## Priority 2 — Orchestration and VEQ reasoning

14. **Recon orchestration**
    - Share collection context across planner and workers.
    - Avoid repeating completed work unless evidence is stale or contradictory.
    - Feed results back into the graph with traceable task/worker/agent lineage.
    - Support cancellation, partial completion and retry without duplicate writes.

15. **VEQ intelligence**
    - Define scoring inputs, confidence representation and explanation format using actual normalized observations and provenance.
    - Recommend tools/actions based on capabilities, platform support, measured readiness and policy.
    - Make recommendations explainable and testable; do not treat a score as proof.

16. **Operator intelligence surfaces**
    - Present relationships, coverage gaps, evidence freshness, confidence and recommended next actions.
    - Show tool verification, installation, execution mode, agent requirements and measured performance on marketplace cards and profiles.
    - Make it possible to inspect why an asset, tool match or recommendation exists.

## Supporting execution-safety requirements

- Shared catalogue visibility never grants execution permission.
- Keep discovery/read-only permissions distinct from active scanning or other execution permissions.
- Any active external action must pass ARTEX target/scope, approval, permission, egress and audit controls.
- Store credentials securely and redact secrets from API responses, logs and marketplace projections.
- Add timeout, cancellation, retry and failure handling.
- Verify third-party interfaces against actual versions; never invent endpoints, ports or schemas.

## Validation gates

- Registry tests prove every agent receives the same current catalogue while retaining policy-specific permissions.
- Discovery tests cover missing executables, version parsing, duplicate entries and changed/removed tools.
- Verification tests distinguish installed from interface-verified and operational.
- Agent matching tests cover CLI-only, GUI/screenshot-capable, browser and API/MCP agents.
- Performance tests record measurements and never infer speed from execution mode alone.
- Parser and normalization fixtures cover valid, malformed, truncated, duplicate and partial results.
- Correlation tests preserve source lineage and conflicting observations.
- Coverage tests distinguish unattempted, failed, negative and positive observations.
- Recommendation tests prove scope/permission/approval constraints are respected.
- End-to-end tests trace tool discovery/selection/execution through ingestion, persistence, graph correlation and operator presentation.
- Run Go/frontend checks and migration tests; record actual results rather than inferring success from committed code.

## Implementation status vocabulary

- **Listed**: metadata exists in the catalogue.
- **Installed**: executable/application presence was detected in a specified environment.
- **Verified**: the actual interface/version was tested successfully.
- **Approved**: an operator has authorized the relevant use; this is not proof of installation or functionality.
- **Operational**: configured and successfully exercised end-to-end in the target environment.
- **Measured**: performance/reliability claims are backed by recorded observations.
