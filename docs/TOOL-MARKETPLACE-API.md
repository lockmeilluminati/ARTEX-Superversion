# Tool Marketplace API

## Endpoint

`GET /api/tool-registry`

Requires the normal ARTEX authenticated API session.

Returns a read-only catalogue projection:

```json
{
  "entries": [
    {
      "id": "runtime.example_tool",
      "name": "example_tool",
      "source": "native",
      "description": "Tool description",
      "enabled": true,
      "kind": "builtin",
      "agents": ["worker"],
      "calls": 0
    },
    {
      "id": "external.yakit",
      "name": "Yakit",
      "source": "external",
      "description": "External tool profile",
      "enabled": false,
      "kind": "external_adapter",
      "manifest": {
        "id": "external.yakit",
        "lifecycle": "candidate",
        "availability": "unknown",
        "trust": "unknown",
        "approval_required": true,
        "verified": false
      }
    }
  ]
}
```

The response combines:

- `native`: code-backed ARTEX tools projected from the existing `tools` table.
- `custom`: operator-defined tools projected from the same table.
- `external`: descriptive manifests persisted in `tool_registry`.

## Security and execution boundary

This endpoint is catalogue metadata only. It does not install tools, create handlers, enable external execution, or bypass ARTEX's existing target, command, approval, and egress controls.

Custom-tool execution specifications and secrets are deliberately excluded from this projection. Existing `GET /api/tools` and the established custom-tool management endpoints remain responsible for the legacy editor workflows.

An external record must move through separate verification, installation/configuration, approval, and adapter implementation work before it can become executable. A manifest's `enabled` field is descriptive state, not an execution grant.

## Current frontend integration

The Tools page loads the unified catalogue and displays registered external entries in the External ecosystem tab. Native and custom tools continue to use the existing API and editors so this migration does not break their current management path.

## Next implementation steps

1. Replace the three source-specific tabs with one searchable catalogue and source/category filters.
2. Add rich per-tool profiles: artwork, specialties, platform compatibility, dependencies, provenance, and documentation.
3. Add operator-managed selection collections that do not imply execution permission.
4. Add explicit verification and installation workflows with audit records.
5. Connect verified adapters to the existing guarded execution pipeline.
