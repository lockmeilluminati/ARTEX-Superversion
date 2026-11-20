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

This commit adds the shared metadata contract, deterministic matching logic, tests and marketplace display fields. The common agent-augmentation hook now derives a conservative runtime profile from assembled default tools and connected MCP servers, persists it in `agent_capability_profiles`, and exposes per-agent compatibility results in `shared_tool_catalogue`. The catalogue remains descriptive: external profiles are not callable, and compatibility never changes tool bindings or grants execution. Profiles refresh when an agent's tools are assembled; they are not a live hardware/OS probe and do not claim GUI support unless the configured browser MCP returns recognized screenshot and interaction tools.

Yakit remains explicitly unknown/unverified until ARTEX checks the actual installed version and supported interface. No assumed endpoint, port, API or GUI automation is declared.


## Shared runtime catalogue visibility

All agents now receive the `shared_tool_catalogue` read-only tool through the common agent augmentation hook. It reads current native/custom runtime tool rows and external registry manifests at call time, so registry metadata changes are visible without rebuilding an agent-specific catalogue. The response intentionally omits custom execution specifications, secrets and input schemas. It labels unknown readiness conservatively and states that catalogue visibility is not a callable-tool grant. Actual execution still depends on the existing per-agent tool bindings, enabled state, handler availability, scope/approval checks and other runtime policy gates. This delivers runtime catalogue visibility, not external-tool execution or automatic agent capability matching.


## Runtime capability observation and matching

Each agent assembly records its observed capability profile with evidence labels and timestamp. `terminal` is inferred only when the assembled baseline contains `Bash`; the platform is the ARTEX runtime's `GOOS` for that terminal environment. `mcp` is recorded only for connected MCP servers that returned tools. Browser, visual-interaction and screen-capture capabilities are added only when the connected `browser` MCP returns recognized interaction/screenshot tool names. API capability is deliberately not inferred from generic HTTP or model reasoning. The latest profile is persisted per agent and used by `shared_tool_catalogue` to attach `toolregistry.MatchAgent` results to external profiles. Unknown metadata remains conditional; incompatible profiles are not presented as executable. This is planning guidance, not an external execution path or automatic permission change.


## Operator capability view

The authenticated `GET /api/agent-capabilities` endpoint returns the latest persisted snapshots without execution configuration or secrets. The System → Tools → Agent capabilities tab shows each configured agent's observed capabilities, runtime platforms, timestamp and evidence labels. Agents without a snapshot are explicitly shown as not observed; the UI does not infer missing capabilities. The view is descriptive and does not alter tool bindings or permissions.


Planner guidance is explicit: use `compatible` as a planning filter, treat `conditional` as requiring operator/environment verification, and treat `incompatible` as unsuitable for that agent. Compatibility is then checked separately from installation, verification, enablement, approval and agent binding. A catalogue entry is callable only when it is explicitly present in the agent's actual callable-tool list.
