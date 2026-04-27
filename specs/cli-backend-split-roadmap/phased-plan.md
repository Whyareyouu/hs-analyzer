# Phased Implementation Plan

## Phase 1: Contracts And Ownership

- Create `pkg/contracts/matchupload`
- Define `match_upload/v1`
- Define API error contract
- Add contract tests and fixtures
- Update existing specs to reference backend ownership where needed

## Phase 2: CLI Match Package And Delivery Boundary

- Build completed match package after session finalization
- Introduce uploader interface
- Replace runtime-only match persister with delivery-oriented persister
- Add tests for `GameEnd -> match package -> uploader`
- Keep existing live card stats snapshot behavior unchanged

## Phase 3: Durable Client Delivery Buffer

- Add local spool
- Add retry and idempotency handling
- Add restart recovery for pending uploads
- Add status output for delivery health
- Add recovery tests for server unavailable and client restart

## Phase 4: Backend Ingestion Service

- Create `services/hs-analyzer-be`
- Add `POST /api/v1/matches`
- Validate `match_upload/v1`
- Implement idempotency behavior
- Add API contract and integration tests

## Phase 5: Server Canonical Storage

- Add Postgres storage layer
- Add migrations on backend side
- Store matches and events
- Add deterministic readback tests
- Keep CLI independent from server schema

## Phase 6: Server Projections And Materialized Views

- Build match summary/materialized views from canonical storage
- Add graph projection as server-owned projection tables or equivalent
- Add rebuild path
- Add projection integrity tests

## Phase 7: History, Analytics And Future Best-Move

- Add match history API
- Add match viewer materialization API
- Add analytics dataset foundations
- Add best-move extension points
- Add performance and regression tests

## Rollout Rule

Each phase must keep `go test ./...` passing and must not break existing one-shot analyze or live-monitoring behavior.

