# Match Upload Contract

## Purpose

Define the stable wire contract between `hs-analyzer-cli` and `hs-analyzer-be`.

The backend receives structured match facts, not raw Hearthstone logs. This keeps server storage independent from client log quirks and makes versioning explicit.

## Endpoint

```http
POST /api/v1/matches
Content-Type: application/json
Idempotency-Key: <match_id | game_id | event_stream_hash>
```

## Payload Version

Initial schema version:

```text
match_upload/v1
```

Unknown schema versions must be rejected with `422 validation_failed`.

## Minimal Payload

```json
{
  "schema_version": "match_upload/v1",
  "match_id": "stable-match-id",
  "game_id": "optional-game-id",
  "started_at": "2026-04-27T12:00:00Z",
  "ended_at": "2026-04-27T12:15:00Z",
  "completion_status": "complete",
  "completion_reason": "game_end",
  "source": {
    "client_name": "hs-analyzer-cli",
    "client_version": "dev",
    "logs_fingerprint": "optional"
  },
  "metadata": {
    "game_mode": "standard"
  },
  "events": [],
  "snapshots": []
}
```

## Required Fields

- `schema_version`
- `match_id`
- `completion_status`
- `completion_reason`
- `source.client_name`
- `events`

`started_at`, `ended_at`, `game_id`, `metadata` and `snapshots` are nullable/optional in v1 because current log parsing may not always recover them perfectly.

## Completion Status

Allowed values:

- `complete`
- `incomplete`

Allowed `completion_reason` values should map to the live session detector reasons:

- `game_end`
- `next_game_start`
- `flush`
- `stream_end`

## Events

`events[]` is the canonical uploaded event stream. The exact event payload is versioned inside the contract package and should be derived from existing `game.Event` types.

Raw log lines are not part of the primary contract. They may be added later as debug attachments only if privacy and size rules are defined.

## Idempotency

Preferred key order:

1. `match_id`
2. `game_id`, if reliable
3. deterministic hash of canonical event stream

Backend behavior:

- same key + same payload hash -> return `200 stored` or `202 accepted`
- same key + different payload hash -> return `409 duplicate_conflict`
- missing key -> return `422 validation_failed`

## Responses

```json
{
  "status": "accepted",
  "match_id": "stable-match-id",
  "server_match_id": "optional",
  "message": ""
}
```

Status codes:

- `202 accepted` - payload accepted, storage/projection may continue async
- `200 stored` - payload already stored or stored synchronously
- `409 duplicate_conflict` - idempotency conflict
- `422 validation_failed` - payload is not valid
- `503 unavailable` - retryable server unavailability

## Validation Rules

Backend validates:

- known `schema_version`
- required fields
- allowed enum values
- stable idempotency key
- parseable timestamps when present
- non-empty `events[]` for completed matches
- event payload compatibility with the declared schema version

## Tests

Required acceptance tests:

1. `TestMatchUploadV1AcceptsValidPayload`
2. `TestMatchUploadV1RejectsUnknownSchemaVersion`
3. `TestMatchUploadV1RejectsMissingIdempotencyKey`
4. `TestMatchUploadV1RejectsInvalidCompletionStatus`
5. `TestMatchUploadV1ReturnsStoredForDuplicateSamePayload`
6. `TestMatchUploadV1ReturnsConflictForDuplicateDifferentPayload`
7. `TestCLIBuildsMatchUploadPayloadAfterGameEnd`

