# Server Storage Ownership

## Purpose

All long-term storage and analytics ownership moves to `hs-analyzer-be`.

The CLI may keep operational files for live state and upload delivery. It must not own product storage schema, SQL migrations, graph rebuild, materialized views or analytics datasets.

## Canonical Storage

Recommended v1 primary storage:

```text
Postgres
```

Backend canonical storage owns:

- matches
- events
- turns
- players metadata
- match completion status
- source/client metadata
- payload schema version
- ingestion audit data

## Projection Ownership

Backend owns secondary projections:

- materialized match summary
- materialized match viewer data
- graph projection
- analytics feature tables
- future best-move datasets

Projections are rebuildable from canonical storage.

## Graph Strategy

V1 graph is a server-owned projection layer. It can be implemented as relational projection tables first.

A dedicated graph DB is a future adapter, not a v1 requirement.

Rules:

- graph facts must be derivable from canonical storage
- graph nodes/edges must reference canonical IDs
- graph rebuild must be deterministic
- graph projection failure must not corrupt canonical storage

## Migration Ownership

Backend owns:

- Postgres schema migrations
- projection schema migrations
- backfill jobs
- projection rebuild tooling
- storage compatibility tests

CLI owns no server storage migrations.

## Materialized Views

Backend materialized views serve:

- match history
- match detail page
- turn/event drilldown
- replay/human output reconstruction
- analytics summaries

CLI can still produce local debug output, but it is not canonical product storage.

## API Surface

Initial backend APIs:

- `POST /api/v1/matches`
- `GET /api/v1/matches`
- `GET /api/v1/matches/{match_id}`
- `GET /api/v1/matches/{match_id}/events`
- `GET /api/v1/matches/{match_id}/materialized`

Only `POST /api/v1/matches` is required for the first split milestone.

## Tests

Required acceptance tests:

1. `TestBackendStoresUploadedMatchInPostgres`
2. `TestBackendStoresEventsWithStableOrder`
3. `TestBackendRejectsInvalidPayloadBeforeStorageWrite`
4. `TestBackendIdempotentWriteDoesNotDuplicateMatch`
5. `TestBackendProjectionBuildUsesCanonicalStorage`
6. `TestBackendProjectionRebuildIsDeterministic`
7. `TestBackendMaterializedViewReadsFromCanonicalStorage`

