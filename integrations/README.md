# Integration adapters

Adapters are replaceable connectors to external tools. Health checks are read-only.
Operation execution is intentionally not implemented by the initial scaffold.

- `Adapter`: identity and health contract.
- `Registry`: in-memory registration for early development.
- `HTTPProbe`: bounded loopback-only health probe.
- `Executor`: future operation contract; must be called only after ARTEX scope and approval gates.
- `yakit/`: Yakit connectivity adapter; endpoint must be verified against the installed version.
- `kali/`: fixed allowlist tool discovery; no scans, installs or arbitrary shell.

Do not bypass `intercept`, `guard`, egress controls, or the existing evidence pipeline.
