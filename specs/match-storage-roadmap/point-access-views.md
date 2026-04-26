# Point Access Views: Turn/Event/Explainable (Task 6.3)

## Цель

Определить точечный доступ к ходу, событию и explainable view без materialization всего матча.

## 1. Turn-level доступ

Контракт:

`GetTurnView(match_id, turn_number, mode, options) -> TurnView`

`TurnView` включает:

- `turn` metadata
- `event_list` выбранного хода
- `snapshot_before/after` (если доступно)
- `decision_points` внутри хода (если есть)

## 2. Event-level доступ

Контракт:

`GetEventView(match_id, event_index, mode, options) -> EventView`

`EventView` включает:

- канонический event payload
- resolved refs (`actor`, `target`, `card`, `entity`)
- локальный контекст (`turn_number`, соседние события по окну)

## 3. Explainable view

Контракт:

`GetExplainableView(match_id, turn_number|decision_point_id, options) -> ExplainableView`

`ExplainableView` включает:

- snapshot/context до решения
- candidate actions (если доступны)
- chosen action
- ближайшие outcomes
- ссылки на canonical IDs для трассировки

## 4. Нефункциональные требования

- выборка одного хода не должна требовать полную materialization матча
- выборка одного события должна быть O(1) или близко к O(1) по индексу
- explainable view должен быть детерминирован на фиксированном наборе данных

## 5. Integration tests

1. `TestGetTurnViewWithoutFullMaterialization`
2. `TestGetEventViewWithReferenceResolution`
3. `TestGetExplainableViewContainsDecisionContext`
4. `TestPointAccessDeterministicOutput`
5. `TestPointAccessVisibilityPolicy`

## 6. Критерий завершения 6.3

Шаг `6.3` считается завершенным, когда:

1. Определены контракты turn/event/explainable view.
2. Зафиксировано, что точечный доступ не требует materialization всего матча.
3. Определены integration tests на выборку хода, события и explainable view.
