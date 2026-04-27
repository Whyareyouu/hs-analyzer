# Testing Strategy

## Acceptance Rule

Every implementation task in this roadmap is accepted only when its tests are written and passing.

Code without tests is not accepted. Passing tests without the required behavior are not accepted.

## Test Levels

### Unit Tests

Use for:

- match package builder
- upload DTO validation
- idempotency key generation
- delivery buffer state transitions
- retry classifier
- backend validation rules
- projection builders

### Contract Tests

Use for:

- `match_upload/v1` encoding/decoding
- backend request/response schema
- error format
- idempotency behavior
- backward-compatible schema evolution

### Integration Tests

Use for:

- CLI finalized match -> uploader call
- uploader -> backend API
- backend API -> Postgres write
- backend storage -> materialized view
- canonical storage -> projection rebuild

### Recovery Tests

Use for:

- CLI crash after spool write
- CLI crash after successful server write but before delivered marker
- server unavailable during upload
- retry after restart
- duplicate delivery after restart

### Regression Tests

Use for:

- live monitor continues when backend is down
- Electron live snapshot continues without server
- existing parser/analyzer behavior remains stable
- `GameEnd` finalization still emits upload package

### Performance Tests

Use for:

- upload payload size
- spool queue growth
- backend ingestion latency
- projection rebuild time
- materialized match query latency

## Required Scenario Matrix

1. Complete game, backend available, match stored.
2. Complete game, backend unavailable, match spooled.
3. Client restarts, pending match uploads successfully.
4. Duplicate upload is idempotent.
5. Duplicate key with different payload is rejected.
6. Invalid payload is rejected and not retried forever.
7. Backend stores canonical data before projection build.
8. Projection rebuild produces deterministic output.
9. Live UI keeps working while backend is unavailable.

## CI Gate

Minimum gate for this roadmap:

```powershell
go test ./...
```

When backend storage is implemented, CI must also include storage integration tests. These tests may use testcontainers, a local Postgres service, or a repository-approved equivalent.

## Documentation Tests

Any change to public contract examples in this spec must be reflected in contract fixtures and validation tests.

