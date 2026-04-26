# Materializer/Renderer Contract (Task 6.1)

## Цель

Определить контракт компонента materialization, который собирает человекочитаемые и debug-представления из compact canonical storage.

## 1. Роли компонентов

- `Storage Reader`:
  - читает canonical сущности (`match`, `turn`, `event`, refs, snapshots).
- `Materializer`:
  - строит intermediate representation (IR) матча/хода.
- `Renderer`:
  - преобразует IR в нужный формат вывода (`text`, `human JSON`, `replay/debug JSON`).

## 2. Режимы materialization

Поддерживаются режимы:

- `text`
- `json_human`
- `json_replay`

Опционально:

- `json_debug` (расширенная диагностика по флагу)

## 3. API-контракты

## 3.1 Полный матч

`MaterializeMatch(match_id, mode, options) -> RenderedDocument`

Гарантии:

- данные берутся только из canonical storage
- output детерминирован на фиксированной версии storage

## 3.2 Один ход

`MaterializeTurn(match_id, turn_number, mode, options) -> RenderedTurn`

Гарантии:

- без обязательной materialization остальных ходов
- корректная резолюция refs для выбранного хода

## 3.3 Один event

`MaterializeEvent(match_id, event_index, mode, options) -> RenderedEvent`

Гарантии:

- наличие минимального контекста (`turn`, `actor/target`, `resolved refs`)

## 4. Опции materialization

Минимальный набор:

- `language` (`ru`, `en`)
- `visibility_scope` (`public`, `player_view`)
- `include_debug_fields` (bool)
- `strict_reference_resolution` (bool)

## 5. Latency требования

Базовые целевые ограничения (подлежат уточнению в performance-этапе):

- `MaterializeMatch`: p95 <= 500ms на матч среднего размера
- `MaterializeTurn`: p95 <= 120ms
- `MaterializeEvent`: p95 <= 60ms

Если цели не выполняются:

- разрешается degrade в debug-полях, но не в корректности фактов

## 6. Contract tests между storage reader и materializer

1. `TestMaterializeMatchFromCanonicalStorage`
2. `TestMaterializeTurnWithoutFullMatchLoad`
3. `TestMaterializeEventWithResolvedReferences`
4. `TestMaterializerDeterminismFixedInput`
5. `TestMaterializerVisibilityScope`
6. `TestMaterializerStrictReferenceResolutionMode`

## 7. Критерий завершения 6.1

Шаг `6.1` считается завершенным, когда:

1. Определены роли `storage reader -> materializer -> renderer`.
2. Зафиксированы режимы materialization и API-контракты.
3. Зафиксированы latency expectations.
4. Определены contract tests между storage reader и formatter/materializer.
