# Phased Implementation Plan

## Goal

Разбить реализацию `hs-analyzer-client` на этапы так, чтобы сначала появился безопасный и тестируемый runtime foundation, затем overlay, затем dashboard/history.

## Phase 1: Client foundation and repo layout

Состав:

- создать `apps/hs-analyzer-client`
- зафиксировать Electron main/preload/renderer structure
- поднять `React + Vite + Tailwind + shadcn/ui`
- внедрить design tokens
- зафиксировать FSD boundaries

Test gate:

- repo structure checks
- Electron security config tests
- renderer smoke tests
- token existence tests

## Phase 2: Local live transport contract

Состав:

- зафиксировать `overlay_snapshot/v1`
- определить initial snapshot + realtime updates
- расширить CLI projection под player deck remaining и opponent played cards
- добавить reconnect/degraded semantics

Test gate:

- contract tests
- CLI integration tests
- reconnect tests

## Phase 3: Overlay window runtime

Состав:

- создать overlay window
- реализовать позиционирование относительно Hearthstone
- реализовать passive mode
- реализовать inspect mode
- добавить tooltip layer

Test gate:

- overlay lifecycle tests
- positioning tests
- passive/inspect transition tests
- tooltip interaction tests

## Phase 4: Overlay UI widgets

Состав:

- player deck panel
- opponent played panel
- card row rendering
- zero-count state
- card art fallback behavior

Test gate:

- component tests
- visual regression tests
- e2e tests на draw/play update flow

## Phase 5: Desktop dashboard shell

Состав:

- sidebar со списком колод
- deck detail header
- summary stats
- matchup visualization
- deck cards grid

Test gate:

- page/widget integration tests
- visual regression tests
- keyboard/focus behavior tests

## Phase 6: Backend analytics integration

Состав:

- deck overview API client
- matchup API client
- cards API client
- loading/error/stale-cache states

Test gate:

- backend contract tests
- client integration tests
- offline/degraded state tests

## Phase 7: Hardening and packaging

Состав:

- reconnect hardening
- DPI/multi-resolution support
- long-running overlay stability
- production packaging/distribution

Test gate:

- soak tests
- performance tests
- packaging smoke tests

## Acceptance Rule

Ни один этап не считается принятым без успешно пройденных тестов его test gate.
