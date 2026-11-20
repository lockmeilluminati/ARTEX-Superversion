# Verification matrix

| Area | Unit/contract | Local mock | Real environment | Runtime wiring |
|---|---|---|---|---|
| Integration registry | Required | N/A | N/A | Pending |
| Yakit health | Required | Required | Pending version/interface verification | Pending |
| Yakit result import | Required | Required | Pending native schema verification | Pending |
| Kali discovery | Required | Fake PATH tests planned | Pending laptop inventory | Pending |
| Tool manifest registry | Required | N/A | N/A | In-memory scaffold |
| Provenance verification | Required | Provider fixtures planned | Pending trusted source providers | Pending |
| Repository lifecycle | Required | Provider fixtures planned | Pending API/source verification | Pending |
| Candidate discovery | Required | Provider fixtures planned | Pending | Pending |
| Documentation drafts | Required | Source fixtures planned | Pending | Pending |
| Normalization/evidence | Required | Required | Pending adapter output | Pending |
| VEQ graph/recommendation | Required | Graph fixtures planned | Pending real ARTEX data adapters | Pending |
| Execution profiles | Required | Policy fixtures planned | Not enabled | Must reuse ARTEX approval |

## Laptop smoke test

1. Run `go test ./...`.
2. Run `go vet ./...`.
3. Start ARTEX with integrations disabled and confirm normal startup.
4. Configure only the verified Yakit health URL and repeat the health test.
5. Confirm Kali discovery reports installed/missing tools without running scans.
6. Do not enable execution profiles until scope, approval, isolation and audit integration tests pass.
