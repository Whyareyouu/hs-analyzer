# Client Delivery Buffer

## Purpose

`hs-analyzer-cli` must not lose completed matches when the backend is unavailable. The client does not own canonical storage, but it does own a durable delivery buffer.

This buffer is an operational queue, not source of truth.

## Responsibilities

The delivery buffer handles:

- durable local spool
- pending upload queue
- retry with backoff
- idempotent send
- crash/restart recovery
- diagnostics for rejected or stuck uploads

It does not provide:

- match history product queries
- materialized match views
- graph projections
- analytics datasets

## Spool Model

Each completed match is written to spool before network delivery.

Recommended logical fields:

- `spool_id`
- `match_id`
- `idempotency_key`
- `payload_path`
- `payload_hash`
- `status`
- `attempt_count`
- `last_attempt_at`
- `next_retry_at`
- `last_error`
- `created_at`
- `updated_at`

Allowed statuses:

- `pending`
- `sending`
- `delivered`
- `retry_wait`
- `rejected`

## Delivery Flow

1. Match is finalized by live monitor.
2. CLI builds `match_upload/v1`.
3. CLI writes payload to local spool.
4. CLI attempts `POST /api/v1/matches`.
5. On `200` or `202`, CLI marks item delivered.
6. On retryable error, CLI schedules retry.
7. On validation error, CLI marks item rejected and keeps diagnostics.
8. On restart, CLI reloads all non-delivered spool items.

## Retry Policy

Retryable:

- network timeout
- connection refused
- `429`
- `500`
- `502`
- `503`
- `504`

Non-retryable:

- `400`
- `401`
- `403`
- `409 duplicate_conflict`
- `422 validation_failed`

Backoff policy should be bounded exponential backoff with jitter. The exact timing is implementation detail, but tests must be deterministic via injected clock.

## Idempotency

The delivery buffer must always send `Idempotency-Key`.

If the CLI crashes after the server stores the match but before local delivered status is written, retry must be safe. Backend idempotency makes this safe.

## Local Data Retention

Delivered items may be cleaned after a configured retention window.

Rejected items should be retained longer for diagnostics because they indicate client/server contract mismatch or corrupted payload.

## Observability

CLI status output should distinguish:

- backend unavailable
- upload pending
- retry scheduled
- upload rejected
- upload delivered

Existing monitor storage health can be reinterpreted as delivery health after this split.

## Tests

Required acceptance tests:

1. `TestDeliveryBufferWritesPayloadBeforeUpload`
2. `TestDeliveryBufferRetriesWhenServerUnavailable`
3. `TestDeliveryBufferSurvivesClientRestart`
4. `TestDeliveryBufferClearsDeliveredItemAfterSuccess`
5. `TestDeliveryBufferMarksValidationErrorRejected`
6. `TestDeliveryBufferUsesStableIdempotencyKey`
7. `TestDeliveryBufferDoesNotBlockLiveMonitorWhenServerDown`

