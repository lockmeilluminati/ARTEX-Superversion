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
- `agent/tools_insert.go` implements `insert_assets` and `SetAssetStore`. The handler derives task ID from runtime context (not model input), loads global and task asset-interception rules fail-closed, evaluates each item before writing, and calls the corresponding `AssetStore.Upsert*` method. It returns per-item results/errors, anchors successful assets to the owner exploration node, records task asset source (`agent`) and adds conservative auto-scope entries. Derived side-effect assets are not blindly auto-scoped.

## Scanner-output parsers present in source

- `agent/scan_ingest.go` and `agent/scan_ingest_tool.go` contain deterministic parsers for Nmap XML, Nuclei JSONL, httpx JSONL and ffuf JSON. The ingestion method accepts a file path under the run working directory, rejects path traversal/symlink escape and non-regular files, caps file size at 256 MiB, caps JSONL lines at 64 KiB and caps one file at 50,000 emitted assets/facts.
- Nmap emits IP and service assets plus one host fact. httpx emits HTTP service assets plus one URL fact. ffuf emits service/URL assets plus one fact per result. Nuclei emits a target asset plus an explicitly unverified scanner-result fact; it does not create a confirmed vulnerability finding.
- Parser inputs are typed and validated. Target-controlled text (banners, titles, template names, extracted values) is retained only as inert quoted fact detail; URL userinfo is stripped. Malformed lines are counted/skipped; truncated files preserve complete records and report truncation.
- Scanner assets flow through `insert_assets`; scanner facts flow through `recordScanFacts`, which deduplicates by source + summary + detail within the task exploration graph. Asset upserts provide the asset-level idempotence. The source contains parser and tool-constructor tests, including PostgreSQL-backed ingestion tests, but these tests have not been run in this work session.
- `agent/bbot_ingest.go` and `agent/bbot_ingest_tool.go` contain a typed BBot JSONL parser. Only `scope_distance == 0` events may create assets through `insert_assets`; other or missing scope distances become `third_party_intel` facts and never auto-enter task scope. Only typed fields drive routing; free text is inert quoted detail. Oversized and malformed lines are counted.
- Both `ingest_bbot` and `ingest_scan_output` are included in `ToolSet.WorkerTools()`; `BuiltinToolSeeds()` derives the worker's default managed-tool bindings from that set. Runtime availability still depends on the persisted tool row remaining enabled/bound, and the actual scanner binaries/commands and end-to-end execution have not been exercised here.

## Current classification

| Area | Status | Evidence / limitation |
|---|---|---|
| Managed native tool metadata and agent binding | Implemented | `agent/toolcatalog.go`, `server/assembly.go` |
| Runtime MCP/skill/host tool assembly | Implemented | `server/assembly.go`; connection and tool discovery remain configuration-dependent |
| Shared tool catalogue visibility | Implemented | Read-only metadata; external entries are not callable |
| General exploration fact persistence | Implemented | `record_fact`; not yet a normalized observation ledger |
| Finding persistence and evidence linkage | Implemented | `report_finding`, `db/findings.go` |
| Asset upsert and API persistence | Implemented | `db/assets.go`, `server/assets.go` |
| `insert_assets` handler-to-store lineage | Implemented | `agent/tools_insert.go` → asset interception → typed `AssetStore.Upsert*` → task provenance/auto-scope |
| External scanner binary detection and execution | Not established | Registry profiles are metadata-only; README toolchain references do not prove local installation or invocation |
| Scanner-specific output parsers | Implemented and included in default worker tool set | Nmap XML, Nuclei JSONL, httpx JSONL, ffuf JSON and BBot JSONL parsers are called by worker tools; actual binary availability and end-to-end execution remain unverified |
| Asset/finding UI visibility | Implemented at API/UI level; run-to-view refresh partial | Assets page reads `/api/assets` and has explicit manual refresh; findings flat/group views refresh periodically, while asset-tree view is loaded on entry/filter change. A scanner run is not yet proven to trigger an asset-page refresh automatically |
| Raw output/evidence pointer attached to every normalized asset observation | Not established | Existing facts/findings can hold evidence text; canonical raw-evidence lineage is still a design gap |

## Gaps to close next

1. Inventory which scanner binaries/commands are actually present in the configured worker environment and how the worker invokes them through Bash. Do not infer installation from the parser or catalogue.
2. Run the existing parser and PostgreSQL-backed fixtures, then add end-to-end fixtures that exercise command → output file → worker ingestion tool → persisted assets/facts.
3. Confirm whether scanner completion should trigger an asset-page refresh signal; the current assets page exposes manual refresh, while finding views have their own refresh behavior.
4. Extend parser coverage only for missing high-value formats after the local command inventory.
5. Define the canonical observation contract using the verified pipeline, without duplicating the existing asset, exploration or finding stores.

## Safety and interpretation

- A tool listed in the marketplace is not necessarily installed, verified, enabled, bound or executable.
- A fact is not automatically a verified asset or vulnerability finding.
- An asset's presence does not prove it was tested; absence does not prove a negative result.
- Compatibility, catalogue visibility and collection recommendations never bypass target scope, approval, interception or egress controls.
- This document does not claim a scanner run, database migration, build or automated test was executed.
