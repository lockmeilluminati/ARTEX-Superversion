-- Seed evidence-backed external tool profiles. These records are descriptive only:
-- no installation, enablement, execution permission, or adapter connectivity is granted.
INSERT INTO tool_registry (id, name, manifest) VALUES
(
  'external.nmap', 'Nmap',
  '{
    "id":"external.nmap",
    "name":"Nmap",
    "description":"Network exploration and security auditing tool for host discovery, port scanning, service/version detection and operating-system fingerprinting.",
    "category":"Network analysis",
    "license":"Nmap Public Source License 0.95 (custom; GPLv2-based, not GPLv2-compatible)",
    "maintainer":"Nmap Project",
    "canonical_repository":"https://github.com/nmap/nmap",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/nmap/nmap"},{"kind":"license","url":"https://github.com/nmap/nmap/blob/master/LICENSE"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["host_discovery","port_scanning","service_detection","os_fingerprinting","network_mapping"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
),
(
  'external.nuclei', 'Nuclei',
  '{
    "id":"external.nuclei",
    "name":"Nuclei",
    "description":"Template-driven vulnerability scanner supporting multiple protocols including HTTP, DNS, TCP, SSL and cloud-related checks.",
    "category":"Web security",
    "license":"MIT",
    "maintainer":"ProjectDiscovery",
    "canonical_repository":"https://github.com/projectdiscovery/nuclei",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/projectdiscovery/nuclei"},{"kind":"project_overview","url":"https://projectdiscovery.io/open-source"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["template_based_scanning","vulnerability_detection","misconfiguration_detection","multi_protocol_scanning","jsonl_output"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
),
(
  'external.httpx', 'httpx',
  '{
    "id":"external.httpx",
    "name":"httpx",
    "description":"HTTP probing toolkit for checking URL and host responsiveness and collecting HTTP metadata such as status, title, web server and TLS details.",
    "category":"Web security",
    "license":"MIT",
    "maintainer":"ProjectDiscovery",
    "canonical_repository":"https://github.com/projectdiscovery/httpx",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/projectdiscovery/httpx"},{"kind":"license","url":"https://github.com/projectdiscovery/httpx/blob/dev/LICENSE.md"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["http_probing","status_detection","title_extraction","technology_fingerprinting","tls_metadata","json_output"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
),
(
  'external.katana', 'Katana',
  '{
    "id":"external.katana",
    "name":"Katana",
    "description":"Configurable web crawling and spidering framework with standard and headless modes, JavaScript crawling and scope controls.",
    "category":"Web security",
    "license":"MIT",
    "maintainer":"ProjectDiscovery",
    "canonical_repository":"https://github.com/projectdiscovery/katana",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/projectdiscovery/katana"},{"kind":"license","url":"https://github.com/projectdiscovery/katana/blob/dev/LICENSE.md"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["web_crawling","javascript_crawling","endpoint_discovery","form_discovery","headless_crawling","scope_control"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
),
(
  'external.ffuf', 'ffuf',
  '{
    "id":"external.ffuf",
    "name":"ffuf",
    "description":"Fast web fuzzer for content discovery, virtual-host discovery and parameter or request-body fuzzing.",
    "category":"Web security",
    "license":"MIT",
    "maintainer":"ffuf contributors",
    "canonical_repository":"https://github.com/ffuf/ffuf",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/ffuf/ffuf"},{"kind":"license","url":"https://github.com/ffuf/ffuf/blob/master/LICENSE"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["content_discovery","virtual_host_discovery","parameter_fuzzing","http_fuzzing","recursive_fuzzing"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
),
(
  'external.dalfox', 'Dalfox',
  '{
    "id":"external.dalfox",
    "name":"Dalfox",
    "description":"Parameter-analysis and XSS scanning tool for identifying reflected, stored and DOM-related cross-site scripting issues.",
    "category":"Web security",
    "license":"MIT",
    "maintainer":"hahwul and contributors",
    "canonical_repository":"https://github.com/hahwul/dalfox",
    "sources":[{"kind":"upstream_repository","url":"https://github.com/hahwul/dalfox"},{"kind":"license","url":"https://github.com/hahwul/dalfox/blob/main/LICENSE.txt"}],
    "platforms":["Linux","Windows","macOS"],
    "capabilities":["xss_detection","parameter_analysis","dom_xss_detection","blind_xss_testing","json_output"],
    "permissions":["discover","read_only"],
    "execution_kind":"external_cli",
    "enabled":false,"availability":"unknown","trust":"unknown","approval_required":true,"lifecycle":"candidate","verified":false
  }'::jsonb
)
ON CONFLICT (id) DO NOTHING;
