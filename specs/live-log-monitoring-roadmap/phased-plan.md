# Live Monitoring Phased Plan (Task 12)

## Этап 1

- архитектурная фиксация live-monitoring и связи с match storage
- acceptance tests для архитектурных решений

## Этап 2

- auto-discovery логов
- live reader for growing `Power.log`
- покрытие unit/integration тестами

## Этап 3

- session detector
- post-game finalization pipeline
- edge-case handling and regression tests

## Этап 4

- persistence завершенного матча в canonical storage
- retry/idempotency
- failure-path tests

## Этап 5

- CLI `monitor` режим
- status output
- graceful shutdown + flush behavior

## Этап 6

- observability/health/self-check
- recovery/state-file strategy
- soak/performance verification

## Этап 7

- live card stats projection для клиента
- per-side aggregation (`self` / `opponent`) по сыгранным картам
- lightweight snapshot delivery без full replay materialization
- recovery semantics для mid-match restart
- unit/integration tests на visibility, aggregation и snapshot consistency

## Gate правило

Каждый этап считается завершенным только после прохождения заранее определенного набора тестов (unit/integration/regression/soak/perf в зависимости от этапа).
