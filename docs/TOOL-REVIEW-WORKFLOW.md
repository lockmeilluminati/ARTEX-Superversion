# External tool operator review workflow

The external-tool profile sheet provides an authenticated operator workflow for recording a metadata review.

## Review states

- **Pending** — no completed operator review is recorded.
- **Approved** — the operator has reviewed the catalogue metadata and considers it acceptable as a descriptive profile.
- **Rejected** — the operator has rejected the profile; a reason is required.
- **Return to pending** — reopens a profile for another review and records that decision in history.

A rejection reason is mandatory. Notes are limited to 2,000 characters. The latest decision is shown on the profile and marketplace card; the profile sheet also shows the 100 most recent decisions.

## Authentication and identity

All API routes are wrapped by ARTEX's existing JWT session middleware. ARTEX currently has a single administrator account (ARTEX), rather than a multi-user role system. The review handler independently validates the session and records the JWT subject as the reviewer. Do not describe this as multi-user RBAC. If ARTEX later gains multiple operator identities, add explicit authorization roles and reviewer identity management before exposing review actions to additional account types.

## Hard safety boundary

A review decision changes only the separate review columns and appends a row to tool_registry_reviews in the same database transaction. It does **not** edit the tool manifest, enabled, permissions, execution_kind, lifecycle, trust, or verification fields. In particular:

- Approved does not mean locally installed, detected, connected, verified, or operational.
- Approved does not install binaries or enable an adapter.
- Approved does not grant execute, install, or any other permission.
- Candidate records remain disabled and retain their existing conservative permission lists.
- Rejected records remain visible as rejected metadata; the decision does not delete a record.

The review history is append-only during normal review operations and is retained with its registry entry. It is deleted only if the parent registry entry itself is deleted (foreign-key cascade).

## API

- GET /api/tool-registry includes review_status, review_note, reviewed_by, and reviewed_at for external entries.
- POST /api/tool-registry/{id}/review accepts a status and note, with status pending, approved, or rejected.
- GET /api/tool-registry/{id}/reviews returns up to 100 recent decisions.

All three routes are protected by the existing API session middleware. The review endpoints accept only IDs beginning with external. Migration 0013_tool_registry_reviews.sql adds the state and history table.

## Remaining validation

Apply migration 0013 on both a clean database and an existing database, run the Go and frontend test/build suites, and exercise approve/reject/reopen through an authenticated instance. Until those checks run, this workflow is committed but not integration-tested.