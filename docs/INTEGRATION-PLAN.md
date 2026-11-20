# First integration: Yakit + ARTEX

## Objective

Establish a small, observable, reversible integration between ARTEX-Superversion and a locally
installed Yakit instance. The first milestone is **connection discovery and health reporting**.
Only after the local API/automation interface is confirmed should we add a real operation.

This keeps the first laptop test focused: ARTEX starts, detects whether Yakit is reachable, reports
the integration status, and records enough diagnostics to troubleshoot a failed connection.

## What we know — and what we must verify

The repository already has:
- A Go HTTP backend and a Next.js UI.
- A custom-tool system for command, script, and HTTP tools.
- An MCP HTTP client and configurable MCP server support.
- Scope checks, approval gates, execution controls, and traffic/evidence storage.

We should reuse these existing capabilities rather than introduce a parallel execution engine.

On the test laptop, verify the installed Yakit version and its supported local automation interface:
official API, plugin/automation interface, MCP endpoint, or another documented local interface.
Do not guess ports, endpoints, authentication, or request formats.

## First implementation sequence

### 1. Integration contract

Represent each integration with:
- Stable ID and display name.
- Connection configuration (secrets stored through the existing secret mechanism).
- Health state: unknown, available, unavailable, or misconfigured.
- Detected version and supported capabilities, when the upstream interface exposes them.
- Last successful check and a sanitized diagnostic message.

Keep connection checks separate from tool execution.

### 2. Yakit connection probe

Implement a read-only probe against the verified local interface:
- Configurable address; no hard-coded assumed port.
- Short timeout and bounded response size.
- No redirects to arbitrary hosts.
- Never log credentials or full sensitive response bodies.
- Clear status for not installed, not running, authentication failure, and incompatible API.

The initial probe must not launch scans, modify projects, or execute arbitrary commands.

### 3. First safe round trip

After the interface is verified, add one deliberately narrow operation against a local lab target
or an explicitly authorized test target. Require the existing ARTEX scope and approval controls.
Record the integration name/version, operation, target, timestamps, result status, and evidence
references. Normalize results before they enter ARTEX findings.

### 4. Mock and tests

Add a fake Yakit endpoint for deterministic tests:
- Healthy response.
- Connection refused / timeout.
- Authentication failure.
- Unexpected response schema.
- Oversized response and redirect rejection.

The mock must not claim that a real Yakit installation is connected.

### 5. Kali integration

Treat Kali as a separate adapter family. Begin with read-only discovery of selected installed
tools and their versions. Do not run package installation or arbitrary shell commands as part of
the first integration milestone. Any later execution must use explicit tool definitions, fixed
argument schemas, scope enforcement, approval, and isolation appropriate to the tool.

## Laptop test checklist

1. Record OS, ARTEX commit, Yakit version, and how Yakit is launched.
2. Confirm whether Yakit exposes a documented local API or automation endpoint.
3. Confirm whether ARTEX runs natively or in Docker. A container's `localhost` is not the host
   laptop; the connection address must match the actual deployment topology.
4. Configure the endpoint explicitly and run the read-only health check.
5. Capture sanitized logs and the response shape; do not share API keys, cookies, tokens, or
   private target data.
6. Only then implement and test the first narrow operation.

## Acceptance criteria

- ARTEX starts normally with the integration disabled.
- A missing or stopped Yakit instance does not break ARTEX.
- The UI/API distinguishes disconnected, connected, and misconfigured states.
- Health checks are bounded, read-only, and do not expose secrets.
- Tests run without Yakit installed.
- No tool executes against a target unless the ARTEX authorization scope and approval policy allow it.

## Design rule

**Connect first. Verify second. Execute third.** Discovery is not authorization, and a successful
connection is not permission to test a target.
