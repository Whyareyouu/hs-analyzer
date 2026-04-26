# Live Monitoring Testing Strategy (Task 10)

## Цель

Зафиксировать обязательный тестовый набор для приемки live-monitoring подсистемы.

## 1. Unit tests

- path discovery/normalization (`internal/config`)
- live reader behavior (grow/truncate/recreate/missing)
- session detector transitions
- post-game finalizer and retry/idempotency
- observability/health transitions

## 2. Integration tests

Ключевой pipeline:

`log append -> parser ParseLine -> session detector -> post-game finalize -> persistence`

Минимальные сценарии:

- один матч complete
- incomplete flush on shutdown
- несколько матчей подряд без рестарта monitor

## 3. Regression tests

- restart recovery via state-file offsets
- behavior with missing/recreated Power.log
- no duplicate persistence on retry/restart
- CLI surface regression for `monitor` flags

## 4. Soak tests

- длительный прогон monitor на серии логов
- проверка отсутствия memory leak и деградации throughput
- проверка стабильности retry/pending queue

## 5. Acceptance gate

Задача live-monitoring не принимается, если:

- отсутствуют тесты на новый runtime-контракт
- падают unit/integration/regression тесты
- soak/perf критерии не подтверждены для релевантного этапа
