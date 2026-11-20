# Curated external tool catalogue profiles

Migration `0011_external_tool_catalog.sql` adds metadata-only profiles for Nmap, Nuclei, httpx, Katana, ffuf and Dalfox.

## Catalogue safety boundary

Every seeded record is deliberately a candidate:
- `enabled=false`, `availability=unknown`, `trust=unknown`, `verified=false`
- `approval_required=true`
- `permissions` contain discovery/read-only descriptors only; no execute or install permission is granted
- `execution_kind=external_cli` describes the expected integration shape, not an implemented runner
- No binary is installed, detected, invoked or connected by this migration

The upstream project and license sources are linked in each manifest. Platform labels describe upstream distribution/support at a high level and do not confirm that a binary exists on the ARTEX host.

## Profiles

| Tool | Category | License recorded | Primary catalogue capabilities |
|---|---|---|---|
| Nmap | Network analysis | Nmap Public Source License 0.95 (custom; GPLv2-based, not GPLv2-compatible) | Host discovery, port scanning, service detection |
| Nuclei | Web security | MIT | Template-based, multi-protocol vulnerability scanning |
| httpx | Web security | MIT | HTTP probing and metadata collection |
| Katana | Web security | MIT | Web crawling, JavaScript crawling, endpoint discovery |
| ffuf | Web security | MIT | Content, virtual-host and parameter fuzzing |
| Dalfox | Web security | MIT | XSS and parameter analysis |

## Evidence links

- [Nmap repository and licensing note](https://github.com/nmap/nmap) · [Nmap license](https://github.com/nmap/nmap/blob/master/LICENSE)
- [Nuclei repository](https://github.com/projectdiscovery/nuclei) · [ProjectDiscovery tools overview](https://projectdiscovery.io/open-source)
- [httpx repository](https://github.com/projectdiscovery/httpx) · [httpx license](https://github.com/projectdiscovery/httpx/blob/dev/LICENSE.md)
- [Katana repository](https://github.com/projectdiscovery/katana) · [Katana license](https://github.com/projectdiscovery/katana/blob/dev/LICENSE.md)
- [ffuf repository](https://github.com/ffuf/ffuf) · [ffuf license](https://github.com/ffuf/ffuf/blob/master/LICENSE)
- [Dalfox repository](https://github.com/hahwul/dalfox) · [Dalfox license](https://github.com/hahwul/dalfox/blob/main/LICENSE.txt)

## Remaining operator work

Before promoting any profile from candidate/unverified:
1. Confirm the actual installed version and executable path on the ARTEX host.
2. Confirm the platform-specific installation source and binary provenance.
3. Add a reviewed adapter with explicit input/output schemas.
4. Connect execution to target/scope gates, approval, rate limits, cancellation and audit/provenance.
5. Run controlled tests against an explicitly authorized local lab target.
6. Only then update availability, trust, lifecycle and verification metadata.
