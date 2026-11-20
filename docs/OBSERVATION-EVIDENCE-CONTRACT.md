# Canonical Observation and Evidence Contract

Status: proposed shared contract; code-level validation and identity helpers implemented. No persistence migration is applied by this document.

## Purpose and boundaries

An observation is a source-attributed claim about an entity at a point in time. It is not automatically a verified fact, an asset, or a vulnerability finding. This contract standardizes collector output while retaining ARTEX's existing stores:

- **Assets** remain the canonical entity inventory in `assets` / `db.AssetStore`.
- **Exploration facts** remain task-scoped conclusions and negative/positive reasoning in the exploration graph.
- **Findings** remain vulnerability records in `findings` / `db.DBFinding`, with their existing verification and report states.
- **Raw evidence** remains in the existing evidence/artifact storage path; observations carry a reference and digest, not a second copy of the artifact.

The `normalizer.Observation` type is an adapter-neutral contract. It is not itself a new database or graph.

## Schema v1

| Field | Type | Required | Meaning |
|---|---|---:|---|
| observation_id | stable string | derived | Exact replay identity for a source claim |
| identity_key | stable string | derived | Entity grouping key; shared by competing claims about the same entity |
| type | enum | yes | domain, host, ip, port, service, url, technology, finding |
| target | string | yes | Normalized entity locator; for port, target is the IP |
| port | uint16 | port observations | 1–65535 |
| protocol | string | no | Transport/application protocol |
| name / value | string | type-dependent | Typed label and observed claim |
| attributes | string map | no | Bounded adapter-owned typed attributes |
| source_tool_id | string | yes | Stable catalogue/tool identifier |
| source_tool_version | string | no | Version actually observed for this run |
| source_run_id | string | yes | Unique execution/run identifier |
| task_id | positive integer | yes | ARTEX task that owns this collection |
| observed_at | RFC3339 timestamp | yes | When the source observation was made |
| evidence | array | yes | At least one raw evidence reference for durable scanner observations |
| scope | enum | yes | unknown, in_scope, out_of_scope |
| confidence | enum | yes | unknown, inferred, low, medium, high |
| verification | enum | yes | not_assessed, unverified, verified, false_positive, blocked |
| collection_status | enum | yes | unknown, not_tested, tested_positive, tested_negative, partial, failed |

The Go validator currently requires the provenance and state fields, and validates evidence entries when supplied. Production adapter acceptance should require at least one evidence reference for scanner-originated observations; operator-authored facts may use a separately documented evidence policy.

### Semantics that must not be collapsed

- `unknown` means the collector/operator does not know the state.
- `not_tested` means no test was performed. It is not a negative result.
- `tested_negative` means a defined test ran and returned a negative result.
- `partial` means the collection ran but its output is incomplete or truncated.
- `failed` means the collection attempt failed; it does not imply the target is absent.
- `inferred` confidence is not equivalent to verified.
- `verified` applies only after an explicit verification process; a scanner match alone remains unverified unless the tool-specific verification contract says otherwise.
- `out_of_scope` observations may be retained as third-party intelligence, but must never be auto-added to task scope.
- A missing state is invalid. Use the explicit `unknown` / `not_assessed` enum values instead.

## Identity, deduplication and conflict handling

`normalizer.NormalizeIdentity` generates an `obsent:` SHA-256 key from entity type and canonical locator (plus port/protocol/name where these define the entity). Domain/host values are lower-cased and trailing dots removed; IP values use `net.IP.String()`; HTTP(S) URLs are normalized to lower-case scheme/host and have userinfo removed.

`ObservationID` is a second SHA-256 key over the identity key, source tool/version, run, task, timestamp, claim value/attributes and state labels. Replaying the same source claim produces the same ID. A new run or different claim produces a new ID.

Do not use an upsert that overwrites a previous observation merely because `identity_key` matches. Keep the latest entity projection in existing asset rows, while retaining distinct source claims and contradictory values in evidence/provenance history. A positive observation from one source and a negative observation from another must both remain inspectable.

## Existing-store mapping

| Contract observation | Existing ARTEX destination |
|---|---|
| domain / host | `assets` root_domain or subdomain; preserve original claim and source in provenance |
| ip | `assets` IP asset |
| port | IP asset open_ports and/or service asset, according to current AssetStore model |
| service | `assets` service |
| url | service or endpoint asset, according to whether it is a base service or discovered route |
| technology | service asset `technologies`; source-specific claim retained in provenance |
| finding | `findings` row only after finding policy/verification gates; otherwise exploration fact with unverified state |
| tested_negative / not_tested / partial / failed | exploration fact / collection coverage state; never manufacture an asset or finding from status alone |
| raw evidence | existing evidence store; reference and digest attached to observation/finding/fact |

Mapping must preserve task ID, source tool, run ID, observed time, confidence, verification and scope. Existing `assets.task_source` / `task_source_summary`, exploration fact source/confidence and finding evidence/counterevidence fields are compatibility targets, but do not currently represent every contract field. Do not silently squeeze run ID or raw evidence into free-text summaries.

## Representative adapter examples

These are normalized contract examples, not claims that a live scan was run.

### Nmap XML: open TCP port

```json
{
  "type": "port",
  "target": "192.0.2.10",
  "port": 443,
  "protocol": "tcp",
  "source_tool_id": "external.nmap",
  "source_tool_version": "observed-version",
  "source_run_id": "run-nmap-001",
  "task_id": 42,
  "observed_at": "2026-11-20T12:00:00Z",
  "evidence": [{"kind":"raw_output","uri":"evidence://run-nmap-001/nmap.xml","digest":"sha256:...","observed_at":"2026-11-20T12:00:00Z"}],
  "scope": "in_scope",
  "confidence": "high",
  "verification": "unverified",
  "collection_status": "tested_positive"
}
```

### httpx JSONL: HTTP service

```json
{
  "type": "service",
  "target": "https://app.example.test",
  "name": "http",
  "value": "status=200",
  "source_tool_id": "external.httpx",
  "source_run_id": "run-httpx-001",
  "task_id": 42,
  "observed_at": "2026-11-20T12:03:00Z",
  "evidence": [{"kind":"raw_output","uri":"evidence://run-httpx-001/httpx.jsonl","observed_at":"2026-11-20T12:03:00Z"}],
  "scope": "in_scope",
  "confidence": "high",
  "verification": "unverified",
  "collection_status": "tested_positive"
}
```

### Nuclei JSONL: candidate finding, not confirmed

```json
{
  "type": "finding",
  "target": "https://app.example.test/login",
  "name": "candidate-template-id",
  "value": "scanner match awaiting independent validation",
  "source_tool_id": "external.nuclei",
  "source_run_id": "run-nuclei-001",
  "task_id": 42,
  "observed_at": "2026-11-20T12:04:00Z",
  "evidence": [{"kind":"raw_output","uri":"evidence://run-nuclei-001/nuclei.jsonl","observed_at":"2026-11-20T12:04:00Z"}],
  "scope": "in_scope",
  "confidence": "medium",
  "verification": "unverified",
  "collection_status": "tested_positive"
}
```

### BBot JSONL: out-of-scope third-party intelligence

```json
{
  "type": "domain",
  "target": "related.example.test",
  "source_tool_id": "external.bbot",
  "source_run_id": "run-bbot-001",
  "task_id": 42,
  "observed_at": "2026-11-20T12:05:00Z",
  "evidence": [{"kind":"raw_output","uri":"evidence://run-bbot-001/output.json","observed_at":"2026-11-20T12:05:00Z"}],
  "scope": "out_of_scope",
  "confidence": "medium",
  "verification": "unverified",
  "collection_status": "tested_positive"
}
```

## Persistence / migration plan (not yet applied)

1. **Compatibility first:** map each observation to the existing asset/fact/finding row and retain current APIs and UI behavior. Do not create a second asset table or graph.
2. **Evidence lineage:** extend the existing evidence/provenance model with a small append-only observation-claim relation keyed by `observation_id`, with optional nullable `asset_id`, `finding_id`, `exploration_node_id`, and task/run/source references. This stores claim history and links into existing records; it is not an independent entity store.
3. **Migration shape:** create a migration only after schema review. Proposed relation fields: `observation_id` primary key, `identity_key` indexed (not unique), `observation_type`, `target`, typed locator columns, `source_tool_id`, `source_tool_version`, `source_run_id`, `task_id`, `observed_at`, `evidence_ref`, `evidence_digest`, `scope_status`, `confidence`, `verification_status`, `collection_status`, `claim_value`, `attributes` JSONB, and nullable links to existing records. `identity_key` must not be unique so conflicts survive.
4. **Backfill:** do not invent source run IDs, evidence references, or timestamps for historical rows. Mark unavailable provenance as unknown and record the backfill origin explicitly.
5. **Rollout:** dual-write from one adapter behind a feature flag, compare old/new projections, add replay/conflict tests, then expand to Nmap/httpx/Nuclei/BBot. Do not switch readers until reconciliation and rollback are proven.

## Contract test requirements

- Missing source tool/run/task/timestamp is rejected.
- Explicit unknown states are valid; empty/unknown is not treated as negative.
- Domain/IP normalization yields stable identity keys.
- Exact replay yields the same observation ID.
- A changed claim or new source run shares the entity key but gets a distinct observation ID.
- Out-of-scope evidence is retained but never auto-scoped.
- Nuclei candidate output remains unverified until an explicit verifier succeeds.
- Partial/truncated and failed collection states remain distinct from tested-negative.
