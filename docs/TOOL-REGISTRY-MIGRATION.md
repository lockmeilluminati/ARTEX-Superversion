# Unified tool catalogue: first migration slice

The first registry importer projects existing ARTEX rows from the `tools` table into namespaced
manifests. It does not modify the database, replace handlers, or change which agents can call tools.

## Stable identity

- Native ARTEX handlers: `artex.native.<existing-key>`
- Operator-defined custom tools: `artex.custom.<existing-key>`
- `legacy_key` preserves the exact database key used by the current runtime.

The registry identity is metadata identity; ARTEX continues resolving actual handlers using the
legacy key and its existing code/DB mechanisms.

## State is deliberately multidimensional

- `lifecycle`: catalogue lifecycle.
- `availability`: whether ARTEX has registered/observed the integration.
- `trust`: why the metadata is trusted (or not).
- `enabled` and `agent_bindings`: existing ARTEX tool-table controls.
- `approval_required`: conservative execution policy marker.

A registered manifest is not proof that a binary is installed, an external service is connected,
or an operation is authorized. Custom tools are operator-configured, not externally verified.
Shell-kind custom tools are represented as declarations, not executable adapters.

## Next integration points

1. Project MCP server tools and skills into the same manifest vocabulary without pretending they
   are rows in the tools table.
2. Add a read-only registry view/API backed by `db.ListTools` and live MCP/skill observations.
3. Add persistence only after the metadata schema and migration semantics are reviewed.
4. Keep execution in existing ARTEX handlers and pass all operations through existing scope,
   permission, approval, egress and evidence controls.


## External tool records

External tools live in the separate `tool_registry` table; they are not fabricated as rows in
the legacy `tools` table and do not receive ARTEX agent bindings automatically. Migration
`0009_tool_registry.sql` creates this table and seeds the first record, `external.yakit`.
The JSON manifest is descriptive catalogue metadata only. It is not an executable handler.

The database access layer exposes `UpsertRegistryEntry` and `ListRegistryEntries`. The
Yakit profile is defined by `toolregistry.YakitManifest` and can be persisted explicitly with
`toolregistry.SeedYakit`; the schema migration also seeds it for fresh/upgraded databases.
