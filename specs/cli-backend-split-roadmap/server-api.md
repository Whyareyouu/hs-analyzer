# Server API

## Purpose

Define the initial backend API surface for `hs-analyzer-be`.

Only match ingestion is required for the first rollout. Read APIs are included to define direction for history, materialized views and analytics.

## Required V1 Endpoint

```http
POST /api/v1/matches
```

Responsibilities:

- validate `match_upload/v1`
- enforce idempotency
- write canonical match data
- trigger or schedule projections
- return stable response/error payload

## Future Read Endpoints

```http
GET /api/v1/matches
GET /api/v1/matches/{match_id}
GET /api/v1/matches/{match_id}/events
GET /api/v1/matches/{match_id}/materialized
GET /api/v1/matches/{match_id}/analytics
```

These endpoints are backend-owned and must read from canonical storage or server-side materialized/projection layers.

## Error Format

Recommended error payload:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "payload is invalid",
    "details": []
  }
}
```

Stable error codes:

- `validation_failed`
- `duplicate_conflict`
- `unauthorized`
- `forbidden`
- `rate_limited`
- `server_unavailable`
- `internal_error`

## Tests

Required acceptance tests:

1. `TestPostMatchesStoresValidPayload`
2. `TestPostMatchesRejectsInvalidPayload`
3. `TestPostMatchesHandlesDuplicateSamePayload`
4. `TestPostMatchesRejectsDuplicateDifferentPayload`
5. `TestPostMatchesReturnsStableErrorShape`
6. `TestBackendReadEndpointsUseServerStorage`

