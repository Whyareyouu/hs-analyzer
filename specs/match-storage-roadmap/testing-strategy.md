# Testing Strategy for Match Storage Subsystem (Task 10)

## Цель

Зафиксировать обязательный test suite для приемки каждой завершенной реализации подсистемы.

## 1. Unit tests

Обязательные группы:

- storage schema contracts
- writer behavior (transaction/idempotency/retry)
- reader contracts
- materializer/renderer contracts
- replay/state reconstruction units

## 2. Integration tests

Ключевой e2e путь:

`logs -> analyzer -> storage writer -> canonical storage -> materializer -> output`

Проверяется:

- корректность фактов
- согласованность turn/event boundaries
- корректность output в `text/human/replay` режимах

## 3. Regression tests

Обязательные регрессии:

- эквивалентность с legacy timeline semantics
- стабильность schema/version migration поведения
- стабильность decision point extraction

## 4. Property/Invariant tests

Ключевые инварианты:

- целостность refs (`event -> turn -> match`, `entity/card refs`)
- детерминированность reconstruction
- отсутствие orphan nodes/edges в projections
- monotonic event order внутри матча

## 5. Test gates (обязательные условия приемки)

Реализация не принимается, если:

- не пройдены обязательные unit/integration/regression/invariant tests
- отсутствуют тесты для нового контракта/схемы
- есть known flaky тесты без стабилизации или quarantine policy

## 6. Рекомендуемая структура тестовых наборов

- `tests/unit/storage/...`
- `tests/unit/materializer/...`
- `tests/integration/pipeline_storage_materialization/...`
- `tests/regression/compatibility/...`
- `tests/invariant/reconstruction/...`

## 7. Критерий завершения 10

Шаг `10` считается завершенным, когда:

1. Зафиксированы unit/integration/regression/invariant test классы.
2. Зафиксирован обязательный e2e pipeline test.
3. Зафиксированы acceptance gates, без которых этап не принимается.
