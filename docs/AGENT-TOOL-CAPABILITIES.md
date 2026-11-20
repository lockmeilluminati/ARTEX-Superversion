# Shared Agent Tool Capability Contract

The tool registry is the shared catalogue for native, custom, external and discovered tools. The existing `GET /api/tool-registry` projection already combines native/custom runtime tools with external registry manifests for the operator marketplace. This contract extends external manifest metadata so the same catalogue can support agent matching.

## Tool readiness metadata

Each manifest may declare:

- `execution_mode`: `cli`, `gui`, `api`, `mcp`, `hybrid` or `unknown`.
- `verification_status`: `unverified`, `partial`, `verified` or `failed`.
- `installation_status`: `unknown`, `installed`, `missing` or `not_applicable`.
- `verification_version` and `verified_at`: the version and time against which the interface was checked.
- `required_agent_capabilities`: explicit interaction requirements.
- `performance`: measured sample count and median startup/run times, with measurement time.

Missing fields are intentionally treated as unknown by clients. Metadata review, installation detection, interface verification, approval and operational status are separate facts.

## Agent capability vocabulary

Current capability identifiers:

- `terminal`: agent can invoke approved command-line tools.
- `screen_capture`: agent can capture the GUI state.
- `visual_interaction`: agent can operate a graphical interface.
- `browser`: agent can interact with a browser.
- `api`: agent can call a supported structured API through an approved adapter.
- `mcp`: agent can use a configured MCP interface.

An agent profile declares the capabilities it actually has, plus supported platforms. Do not grant a capability merely because a model could theoretically reason about it.

## Matching behavior

`toolregistry.MatchAgent` evaluates declared requirements:

- CLI implies `terminal`.
- GUI implies both `screen_capture` and `visual_interaction`.
- API implies `api`.
- MCP implies `mcp`.
- Hybrid tools must list their concrete required capabilities.
- Platform mismatch is incompatible.
- Missing agent or tool metadata produces a conditional result, not an optimistic match.

The result is a compatibility assessment only. It never enables a tool, grants permissions, installs software or bypasses target/scope, approval, egress or audit controls. A compatible agent/tool pair may still be unusable because the tool is missing, unverified, disabled or not approved.

## Current integration boundary

This commit adds the shared metadata contract, deterministic matching logic, tests and marketplace display fields. The existing marketplace API exposes the manifest metadata. It does **not yet** inject the external registry into planner/worker runtime prompts or allow external tools to execute. Agent profiles are not yet persisted or populated from actual runtime capabilities. Those are subsequent integration jobs.

Yakit remains explicitly unknown/unverified until ARTEX checks the actual installed version and supported interface. No assumed endpoint, port, API or GUI automation is declared.
