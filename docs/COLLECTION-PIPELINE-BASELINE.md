# Collection Pipeline Baseline

Last reviewed: 2026-11-20

This is a source-level baseline, not an end-to-end verification. It records what the current default-branch code shows and separates implemented persistence paths from unverified/scaffolded integrations.

## Confirmed runtime assembly and tool visibility

1. `agent/toolcatalog.go` constructs ARTEX's managed domain-tool sets for `mainagent`, `planner`, `worker`, `auto` and `pentest`. It seeds tool metadata (name, description, schema and default agent bindings) into the database.
2. The `ToolResolve` hook filters assembled managed tools against database enablement and per-agent bindings, then applies description/schema overrides and scalar defaults. It does not create a handler for a tool whose implementation is absent.
3. `server/assembly.go` adds the read-only `shared_tool_catalogue`, configured MCP tools, skills and host tools. Runtime capability profiles are observed from assembled defaults and connected MCP tool names. The custom-tool executor is also assembled through the host-tool path.
4. The Norma SDK default tool set is included in the agent baseline. The observed capability logic only infers terminal when `Bash` is present. This is not evidence that any particular scanner binary is installed.

## Confirmed result and persistence paths

### Exploration observations

- `agent/tools.go` exposes `record_fact` for exploration conclusions, including positive and negative results.
- The input contract carries summary, detail, evidence, confidence, source and optional intent/asset references. Source values distinguish target response, tool output, third-party intelligence, operator input and inference.
- Facts are linked into the task exploration graph. This is a general observation/conclusion path, not the canonical normalized asset observation contract proposed in the roadmap.
- Negative conclusions are supported, but the code's existence alone does not prove every scanner result is consistently recorded.

### Findings

- `report_finding` requires task context and records a finding with severity, summary, evidence, worker, intent and optional asset IDs, counterevidence and confidence rationale.
- Traffic evidence can be linked through verified traffic references or a hint. If the traffic evidence store is unavailable when references are supplied, the tool refuses to register the finding.
- The finding path uses the finding recorder when configured, otherwise the exploration store when no traffic refs are supplied. The returned record distinguishes the finding row ID from the exploration finding-node ID.
- `db/findings.go` stores standalone findings, task/node links, asset IDs, evidence, counterevidence, confidence rationale, verification state and report state. Global findings survive task deletion by design.

### Assets

- `db/assets.go` provides persistent upsert paths for root domains, IPs, subdomains, HTTP services, other services and endpoints.
- Upserts merge repeated data, preserve task associations, and refresh `last_seen`; subdomain A/AAAA records can also create/link IP assets. Asset records include technologies, HTTP metadata, ports, endpoint parameters and an `extra` JSON field.
- IP validation rejects hostnames in the IP field. Domain normalization and URL normalization are performed in the asset store.
- `server/assets.go` exposes `POST /api/assets`, processing a batch item by item. It returns per-item success IDs and per-item errors rather than failing the entire batch. Successful API inserts are tagged with source `api` and a task-level source note when a task ID is supplied.
- `GET /api/assets` supports paging, type, company/task scoping and validated DSL search.
- The managed-tool catalogue seeds `insert_assets` for the Auto and Pentest agents, but this baseline has not yet traced the full tool handler and every call site into the same asset-store path.

## Current classification

| Area | Status | Evidence / limitation |
|---|---|---|
| Managed native tool metadata and agent binding | Implemented | `agent/toolcatalog.go`, `server/assembly.go` |
| Runtime MCP/skill/host tool assembly | Implemented | `server/assembly.go`; connection and tool discovery remain configuration-dependent |
| Shared tool catalogue visibility | Implemented | Read-only metadata; external entries are not callable |
| General exploration fact persistence | Implemented | `record_fact`; not yet a normalized observation ledger |
| Finding persistence and evidence linkage | Implemented | `report_finding`, `db/findings.go` |
| Asset upsert and API persistence | Implemented | `db/assets.go`, `server/assets.go` |
| Full `insert_assets` handler-to-store lineage | Partial / not yet traced | Tool is seeded for Auto/Pentest; handler implementation and all producers still need a focused trace |
| External scanner binary detection and execution | Not established | Registry profiles are metadata-only; README toolchain references do not prove local installation or invocation |
| Scanner-specific output parsers and normalized ingestion | Not established | No complete parser inventory or end-to-end fixture chain has been verified in this baseline |
| Asset/finding UI visibility from a real scanner run | Not yet traced | API routes are confirmed; end-to-end frontend data flow and refresh behavior still need inspection |
| Raw output/evidence pointer attached to every normalized asset observation | Not established | Existing facts/findings can hold evidence text; canonical raw-evidence lineage is still a design gap |

## Gaps to close next

1. Locate and trace the `insert_assets` implementation and every caller; verify validation, partial-batch behavior, provenance and idempotence.
2. Inventory actual scanner commands/tools invoked by Bash, custom command tools, MCP and any dedicated adapters. Distinguish tool documentation from executable integration.
3. Inventory parsers (including any file-based imports such as BBot) and trace each parser output to facts, assets, findings or nowhere.
4. Trace the frontend routes/components consuming `GET /api/assets` and `GET /api/exploration/findings`, including task-scoped views and refresh.
5. Add end-to-end fixtures that start with representative tool output and assert persisted assets/facts/findings plus source lineage.
6. Use those findings to define the canonical observation contract without duplicating the existing asset, exploration or finding stores.

## Safety and interpretation

- A tool listed in the marketplace is not necessarily installed, verified, enabled, bound or executable.
- A fact is not automatically a verified asset or vulnerability finding.
- An asset's presence does not prove it was tested; absence does not prove a negative result.
- Compatibility, catalogue visibility and collection recommendations never bypass target scope, approval, interception or egress controls.
- This document does not claim a scanner run, database migration, build or automated test was executed.
