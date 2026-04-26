# Performance and Footprint Strategy (Task 11)

## Цель

Определить performance и storage-метрики, а также acceptance criteria для новой подсистемы.

## 1. Метрики размера и скорости

### 1.1 Размер хранения

- `bytes_per_match` (small/medium/long match)
- `bytes_per_100_matches`
- `bytes_per_1000_matches`

### 1.2 Запись

- `write_match_p50/p95`
- `write_matches_per_sec`

### 1.3 Materialization

- `materialize_match_p50/p95`
- `materialize_turn_p50/p95`
- `materialize_event_p50/p95`

### 1.4 Точечный доступ

- `read_turn_p50/p95`
- `read_event_p50/p95`

## 2. Сравнение с legacy `timeline_*.json`

Обязательный baseline:

- размер legacy timeline на том же наборе матчей
- latency получения эквивалентного output

Цель:

- существенное снижение storage footprint
- не хуже latency для ключевых сценариев чтения

## 3. Целевые показатели (initial targets)

- уменьшение размера хранения на матч минимум на `40%` относительно legacy heavy JSON baseline
- p95 чтения одного хода <= `120ms`
- p95 materialization одного матча среднего размера <= `500ms`

Точные значения финализируются после первых benchmark прогонов.

## 4. Политика тяжелых debug-представлений

- тяжелые debug поля включаются только explicit флагом
- default path должен оставаться оптимизированным по размеру и latency

## 5. Benchmark tests

1. `BenchmarkWriteMatchSmall/Medium/Long`
2. `BenchmarkReadTurnByMatchTurn`
3. `BenchmarkReadEventByIndex`
4. `BenchmarkMaterializeMatchHuman`
5. `BenchmarkMaterializeMatchReplay`
6. `BenchmarkStorageFootprintVsLegacyTimeline`

## 6. Performance acceptance criteria

Этап принимается, когда:

1. Собран reproducible benchmark report.
2. Достигнуты или аргументированно пересогласованы целевые thresholds.
3. Нет критических регрессий относительно legacy baseline по ключевым user сценариям.

## 7. Критерий завершения 11

Шаг `11` считается завершенным, когда:

1. Зафиксирован набор метрик размера/скорости.
2. Зафиксирована методика сравнения с legacy timeline baseline.
3. Зафиксированы benchmark tests и acceptance criteria.
