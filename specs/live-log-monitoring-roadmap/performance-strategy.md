# Live Monitoring Performance Strategy (Task 11)

## Цель

Определить performance-метрики и критерии приемки для long-running live режима.

## 1. Метрики

- CPU usage monitor процесса (idle / active game)
- memory footprint (steady-state)
- latency обработки новой строки `Power.log`
- latency от завершения матча до успешной persistence
- pending queue size under transient storage failures

## 2. Целевые ориентиры (initial)

- idle CPU: low single-digit %
- bounded memory growth on long run
- строка `Power.log` до обработки: sub-second p95
- завершение матча до persistence: sub-second to low seconds (в зависимости от storage path)

## 3. Benchmark/soak workload

- synthetic append stream (burst + sparse patterns)
- log rotation/recreate scenarios
- storage transient failure window with retries
- long-running session with multiple matches

## 4. Acceptance criteria

- нет неконтролируемого роста памяти
- latency укладывается в согласованные пороги
- retry механизм не приводит к бесконечному росту pending
- observability отражает деградацию/восстановление корректно
