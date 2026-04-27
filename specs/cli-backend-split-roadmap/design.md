# Design Document: CLI / Backend Split

## Overview

Новая архитектура разделяет текущий `hsanalyzer` на два продукта:

- `hs-analyzer-cli` - локальный агент, который работает рядом с Hearthstone и имеет доступ к локальным логам.
- `hs-analyzer-be` - сервер, который принимает завершенные матчи и владеет долговременным хранением, projections и аналитикой.

Это меняет смысл `match-storage-roadmap`: canonical storage больше не должен быть local desktop ownership. Он становится backend ownership, а CLI получает только durable delivery buffer для надежной доставки.

## Target Architecture

```text
Hearthstone Logs
      |
      v
apps/hs-analyzer-cli
  log discovery
  live reader
  parsers
  session detector
  live Electron snapshot
  completed match package
  outbound queue + spool
      |
      | POST /api/v1/matches
      v
apps/hs-analyzer-be
  ingestion API
  payload validation
  idempotency
  canonical Postgres storage
  materialized views
  graph projections
  match history API
  analytics / future best-move
```

## `hs-analyzer-cli` Responsibilities

CLI keeps all logic that depends on local files or live local runtime:

- Hearthstone logs auto-discovery
- `Power.log` / `LoadingScreen.log` live reader
- parsers and event stream generation
- live session detector
- post-game match package assembly
- live card stats snapshot for Electron
- local status/observability for watcher, parser and delivery
- outbound delivery queue
- local spool for pending match uploads
- retry/idempotent upload to backend

CLI must not own:

- server canonical DB schema
- SQL migrations
- graph DB schema
- graph/materialized projection rebuild
- match history query layer
- heavy analytics and best-move

## `hs-analyzer-be` Responsibilities

Backend owns durable product data and expensive processing:

- `POST /api/v1/matches`
- payload schema validation
- idempotency and duplicate detection
- canonical match storage in Postgres
- event and metadata persistence
- server-side materialized views
- graph projection build/rebuild
- history queries
- match viewer materialization
- analytics and future best-move

Backend must not depend on raw Hearthstone logs as the primary ingest format. Its primary input is a versioned match upload payload.

## Shared Contracts

Shared DTOs and protocol constants live in `pkg/contracts`.

Initial packages:

```text
pkg/contracts/
  matchupload/
  live/
  apierrors/
```

`matchupload` defines `match_upload/v1`. CLI uses it to serialize uploads; backend uses it to validate and decode requests.

## Data Flow

1. CLI monitors logs and emits game events.
2. CLI updates local live snapshots for Electron immediately.
3. Session detector closes the active match on `GameEnd` or fallback finalization.
4. CLI builds a completed match package from the session events and reconstructed state.
5. CLI writes package to outbound queue before attempting network delivery.
6. CLI uploads package to `POST /api/v1/matches`.
7. Backend validates schema, idempotency key and payload invariants.
8. Backend writes canonical data to Postgres.
9. Backend builds or schedules materialized/graph projections.
10. CLI marks spool item delivered only after server success.

## Storage Ownership

Postgres is the v1 server canonical storage.

Graph data is a server-owned projection over canonical storage. A dedicated graph DB can be added later as an adapter, but v1 must not require it.

CLI can store delivery spool files, runtime offsets and live snapshot recovery data. These are operational buffers, not product source of truth.

## Integration With Existing Roadmaps

`specs/hearthstone-log-analyzer` remains the base parser/analyzer domain.

`specs/live-log-monitoring-roadmap` remains the base for local live monitoring, session detection and Electron-friendly live projection.

`specs/match-storage-roadmap` remains useful for canonical match structure, projections and materialization, but its implementation ownership moves to `hs-analyzer-be`.

## Non-Goals For V1

- server-side parsing of raw Hearthstone logs
- separate graph DB as required infrastructure
- remote real-time spectator UI
- multi-account authentication model
- cloud deployment details beyond API/storage boundaries

## Completion Criteria

The split architecture is accepted only when:

- CLI can complete a match and enqueue/upload a versioned package
- backend can accept, validate and store that package
- server unavailability does not lose a completed match
- storage/projection code is backend-owned
- tests cover contract, retry, idempotency and recovery behavior
