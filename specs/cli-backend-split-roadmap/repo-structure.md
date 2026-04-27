# Repo Structure

## Purpose

Define the target folder layout for splitting the current application into a CLI ingestion agent and backend service while keeping shared contracts explicit.

## Target Layout

```text
apps/
  hs-analyzer-cli/
    cmd/
    internal/
      analyzer-runtime/
      delivery/
      electron-live/
      live-monitor/
      match-package/
      parser-runtime/
      session/

apps/
  hs-analyzer-be/
    cmd/
    internal/
      analytics/
      api/
      bestmove/
      ingestion/
      materialized/
      matches/
      projections/
      storage/
        postgres/
      validation/

pkg/
  contracts/
    apierrors/
    live/
    matchupload/
```

## Current Code Mapping

Current local app code maps to `apps/hs-analyzer-cli`:

- `cmd/hsanalyzer` -> `apps/hs-analyzer-cli/cmd`
- `internal/parser` -> CLI parser runtime
- `internal/logreader` -> CLI parser/live runtime
- `internal/monitor` -> CLI live monitor/session/delivery boundary
- `internal/analyzer`, `internal/game`, `internal/state`, `internal/card`, `internal/player`, `internal/zone`, `internal/turn`, `internal/entity`, `internal/types` -> CLI domain/runtime foundation
- `internal/formatter` -> local debug/export formatter until backend materialization is implemented
- `internal/carddb` -> CLI card metadata lookup for local enrichment; backend can later own a server-side card dictionary separately

New backend code maps to `apps/hs-analyzer-be`.

Shared upload DTOs and protocol errors map to `pkg/contracts`.

## Ownership Rules

CLI-only packages must not import backend internals.

Backend-only packages must not import CLI runtime internals unless a package has been explicitly promoted to `pkg/contracts` or another shared package.

Contracts should contain DTOs, enum constants and validation helpers that are safe to share. They should not contain parser runtime, storage clients or business workflows.

## Migration Strategy

The move to `apps/` should be phased:

1. Add shared contracts without moving runtime code.
2. Introduce backend service in `apps/hs-analyzer-be`.
3. Introduce CLI uploader and delivery buffer.
4. Move current `cmd/hsanalyzer` into `apps/hs-analyzer-cli`.
5. Move CLI-owned internal packages under `apps/hs-analyzer-cli/internal` only after imports and tests are stable.
6. Keep module imports consistent with `go.mod` during each step.

## Acceptance Tests

Required tests and checks:

1. `TestNoBackendInternalImportsFromCLI`
2. `TestNoCLIInternalImportsFromBackend`
3. `TestSharedContractsHaveNoRuntimeDependencies`
4. `go test ./...`
5. Static import check for forbidden dependency direction
