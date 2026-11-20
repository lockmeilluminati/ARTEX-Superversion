# Yakit tool profile

**Registry ID:** `external.yakit`  
**Catalogue lifecycle:** `candidate`  
**Initial enabled state:** false  
**Verification state:** unverified  
**Execution state:** not executable

## Role in ARTEX

Yakit is the first external application recorded in the Superversion tool catalogue. It is
one integration among many; ARTEX remains the orchestration, authorization, evidence and
audit layer. Yakit is not a replacement for ARTEX's native tools, Kali utilities, MCP tools
or the unified registry.

The catalogue profile describes Yakit as a security-testing platform used for web traffic
inspection, request replay, web security testing and plugin-based workflows. These are
descriptive capability labels, not proof that a particular installed version exposes a
supported ARTEX automation API.

## Source and identity

- Stable ARTEX ID: `external.yakit`
- Upstream project reference: https://github.com/yaklang/yakit
- Source type: upstream project + operator-installed application
- Supported local automation interface: **not yet verified**

## Initial trust and permission posture

| Property | Initial value |
|---|---|
| Lifecycle | candidate |
| Availability | unknown |
| Trust | unknown |
| Verified | false |
| Enabled | false |
| Permissions | discover, read_only |
| Approval required | true |
| Execution kind | external_adapter |

The initial registry entry must not be interpreted as a successful connection, installed
binary detection, API compatibility, or authorization to run a test. No endpoint, port,
authentication method, request schema or response schema is assumed.

## Verification checklist

Before implementing a version-specific adapter, record:

1. Installed Yakit version and build.
2. Officially documented local automation/API mechanism for that version.
3. Bind address and port as configured by the operator.
4. Authentication requirements and credential storage requirements.
5. Read-only health/status request and its documented response.
6. Supported operation names, parameter schemas and result formats.
7. How Yakit traffic and findings can be exported with stable native references.

The first connection test should be read-only and loopback-restricted. Do not infer an API
from an open port or undocumented endpoint. Do not embed credentials in URLs or logs.

## Future integration path

1. Verify the installed version and official interface on the operator's laptop.
2. Configure the local endpoint explicitly and report health without running a test.
3. Add version-specific request/response translation behind the Yakit adapter.
4. Propose one low-impact, authorized lab operation with explicit target scope and approval.
5. Normalize returned findings and traffic references into ARTEX evidence.
6. Promote the registry lifecycle only after the operator verifies the integration.

Yakit's presence in the catalogue does not bypass ARTEX target gates, approval controls,
egress controls, evidence handling or audit logging.
