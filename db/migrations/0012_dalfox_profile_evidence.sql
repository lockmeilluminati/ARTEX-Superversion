-- Refresh Dalfox profile details from its current upstream documentation.
-- This remains metadata only; the profile is intentionally not verified or executable.
UPDATE tool_registry
SET manifest = manifest || '{
  "description":"XSS-focused security scanner with parameter discovery, reflected/stored/DOM-based XSS analysis, AST-based checks and optional out-of-band testing.",
  "capabilities":["parameter_analysis","parameter_mining","reflected_xss_detection","stored_xss_detection","dom_xss_analysis","ast_analysis","blind_xss_testing","json_output"],
  "sources":[
    {"kind":"upstream_repository","url":"https://github.com/hahwul/dalfox"},
    {"kind":"upstream_documentation","url":"https://github.com/hahwul/dalfox/blob/main/README.md"},
    {"kind":"license","url":"https://github.com/hahwul/dalfox/blob/main/LICENSE.txt"}
  ]
}'::jsonb,
updated_at = now()
WHERE id = 'external.dalfox'
  AND COALESCE(manifest->>'verified', 'false') <> 'true';
