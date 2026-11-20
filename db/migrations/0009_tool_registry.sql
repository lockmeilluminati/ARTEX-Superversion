CREATE TABLE IF NOT EXISTS tool_registry (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    manifest    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tool_registry_name_idx ON tool_registry (name);

-- Seed the first external tool profile. ON CONFLICT DO NOTHING preserves any
-- operator-reviewed metadata if this migration is applied again manually.
INSERT INTO tool_registry (id, name, manifest)
VALUES (
    'external.yakit',
    'Yakit',
    '{
      "id": "external.yakit",
      "name": "Yakit",
      "description": "External cyber security all-in-one platform. ARTEX connectivity and supported automation operations must be verified against the installed Yakit version before use.",
      "license": "AGPL-3.0",
      "maintainer": "yaklang",
      "canonical_repository": "https://github.com/yaklang/yakit",
      "sources": [
        {"kind": "upstream_project", "url": "https://github.com/yaklang/yakit"},
        {"kind": "operator_installed_application"}
      ],
      "capabilities": ["traffic_inspection", "request_replay", "web_security_testing", "plugin_workflows"],
      "permissions": ["discover", "read_only"],
      "execution_kind": "external_adapter",
      "enabled": false,
      "availability": "unknown",
      "trust": "unknown",
      "approval_required": true,
      "lifecycle": "candidate",
      "verified": false
    }'::jsonb
)
ON CONFLICT (id) DO NOTHING;
