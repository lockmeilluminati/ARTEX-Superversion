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


## Repeatable operator verification (added 2026-11-20)

Run these checks on the **same host/container that executes ARTEX worker commands**. A developer workstation inventory is not evidence about a remote worker.

1. Inventory candidate binaries without installing or scanning anything:
   `bash scripts/inventory-collection-tools.sh`
   Use `--strict` when the deployment expects every listed candidate to be installed. The script reports command path and the tool-specific version output for Nmap, Nuclei, httpx, Katana, ffuf, Dalfox and BBot. A missing binary is a deployment fact, not a parser failure; do not install tools automatically as part of verification.
2. Run focused parser/normalizer tests: `go test ./normalizer ./agent`. Then run the repository gate with a configured test PostgreSQL service: `bash scripts/verify.sh --go-only`. Record the Go version, PostgreSQL version and test output. Do not report a pass when tests were skipped due to missing configuration.
3. In an explicitly authorized disposable lab, confirm the worker has `ingest_scan_output` and/or `ingest_bbot` enabled and bound. Use each installed scanner's own help/version output to confirm flags before running it. Write output beneath the active task working directory, then invoke the matching ingestion tool:
   - Nmap XML: `-oX <working-dir-relative-file>` → format `nmap-xml`.
   - Nuclei JSONL: `-jsonl -o <working-dir-relative-file>` → format `nuclei-jsonl`.
   - httpx JSON: `-json -o <working-dir-relative-file>` → format `httpx-jsonl`.
   - ffuf JSON: `-of json -o <working-dir-relative-file>` → format `ffuf-json`.
   - BBot: use its configured JSONL output file → `ingest_bbot`.
4. For each exercised format, record the command/tool version, output filename and checksum, ingestion response counts, database asset/fact IDs, task association and scope outcome. Repeat ingestion of the same artifact and confirm replay behavior. Confirm out-of-scope BBot events remain third-party facts and do not enter assets/task scope.
5. In the authenticated ARTEX UI, refresh the Assets view and confirm the persisted records appear. The current UI has manual refresh; automatic scanner-completion refresh is not yet established.

Do not run active scans against public or third-party targets for this check. The command inventory and source-level parser tests are safe preparatory checks; end-to-end verification must use an explicitly authorized lab.

**Evidence record:** no scanner inventory, parser test, PostgreSQL test, live command-to-ingestion run or UI refresh has been executed by this repository-edit session. The helper and procedure make the checks repeatable but do not change the pipeline's operational status.

## Safety and interpretation

- A tool listed in the marketplace is not necessarily installed, verified, enabled, bound or executable.
- A fact is not automatically a verified asset or vulnerability finding.
- An asset's presence does not prove it was tested; absence does not prove a negative result.
- Compatibility, catalogue visibility and collection recommendations never bypass target scope, approval, interception or egress controls.
- This document does not claim a scanner run, database migration, build or automated test was executed.


## Attack-surface relationship correlation (Job 7 initial implementation)

- `assetgraph.Correlate([]db.Asset)` derives a deterministic, read-only edge projection from existing asset records.
- Supported relationships are root-domain membership, subdomain-to-IP resolution, service/endpoint host association, and endpoint-to-HTTP-service origin association.
- Each edge includes source/target asset IDs, a relation label, the exact field-level rationale, confidence and an observed/inferred marker.
- The correlator deliberately does not infer edges from shared company/task membership, subnet proximity or technology strings alone. It does not mutate assets or claim that inferred endpoint/service origin matches are direct observations.
- Current limitation: this is a pure package and is not yet wired to DB loading, findings, evidence claims, an authenticated API or UI. Once Job 9 observation persistence is implemented, edges should reference supporting claim IDs and preserve conflicting claims instead of collapsing them.


## Collection coverage intelligence (Job 6)

- The regular ARTEX program already has an asset coverage denominator/backlog in `db/task_assets_context.go`, and a worker-reported risk-class ledger in `db/coverage_ledger.go` exposed through `record_coverage` and planner wrap-up context.
- The asset coverage signal is intentionally described as coarse: a fact anchor means the asset was investigated, not that every vulnerability or discovery dimension was tested.
- Job 6 adds the explicit `test_failed` outcome, which remains an open gap; it cannot be interpreted as a clean/negative result.
- `AssessCoverageEntry` classifies absent entries as `not_tested`, findings as `tested_positive`, evidence-backed negative entries as `tested_negative`, ruled-out/not-applicable entries with reasons as `not_applicable`, follow-up as `partial`, failed attempts as `failed`, stale prior results as `stale` (retaining the prior state), and missing/invalid provenance as `unknown`.
- The DB ledger read attaches derived assessments using a conservative 30-day default. The pure assessment helper accepts a caller-selected freshness window. No stored timestamp is rewritten.
- Migration `0015_coverage_failure_and_freshness.sql` extends the existing outcome constraint; `db/schema.sql` supports clean database creation. This does not alter normal task execution, asset writes, scope gates, approval or egress.
- Remaining: operator UI/API state filters, dimension-specific evidence beyond fact anchors, durable evidence claim links (Job 9), and executed Go/PostgreSQL/migration tests. No tests or migrations were run in this work session.
