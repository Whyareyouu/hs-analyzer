# Secondary Projection Build Strategy (Task 8.3)

## Цель

Определить, как и когда строятся secondary projections (включая graph projection) из primary storage.

## 1. Модель построения

Базовая стратегия:

- canonical storage write - синхронная обязательная операция
- projection build - асинхронная операция по событию сохранения матча

Опционально:

- синхронный projection build для debug/small batch режимов

## 2. Rebuild стратегия

Поддерживаются:

- `full rebuild` всех projections из canonical storage
- `incremental rebuild` по новым/измененным матчам

Требование:

- full rebuild остается эталонным способом восстановления целостности проекций

## 3. Версионирование projections

- `projection_version` хранится отдельно от `schema_version` canonical storage
- изменение projection schema допускает полную переиндексацию без изменения canonical layer

## 4. Согласованность проекций

Гарантии:

- projection узлы/ребра всегда ссылаются на canonical IDs
- при недоступности projection слоя canonical read/write продолжает работать

## 5. Tests

1. `TestProjectionIncrementalBuildFromCanonical`
2. `TestProjectionFullRebuildFromCanonical`
3. `TestProjectionVersionedReindex`
4. `TestProjectionReferenceConsistency`
5. `TestCanonicalWorksWhenProjectionUnavailable`

## 6. Критерий завершения 8.3

Шаг `8.3` считается завершенным, когда:

1. Определен режим построения projections (sync/async policy).
2. Определены стратегии incremental и full rebuild.
3. Зафиксировано versioning правило для projections.
4. Определены тесты на rebuild, versioning и согласованность.
