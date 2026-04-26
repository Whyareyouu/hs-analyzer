# State Reconstruction Strategy (Task 5.2)

## Цель

Определить детерминированный алгоритм восстановления состояния произвольного хода из `snapshot + subsequent events`.

## 1. Алгоритм реконструкции

Для target позиции (`match_id`, `turn_number`, optional `event_index`):

1. Выбрать ближайший предшествующий snapshot/checkpoint.
2. Загрузить canonical events от snapshot-anchor до target.
3. Применить события в стабильном порядке `event_index ASC`.
4. Получить reconstructed state на target момент.

## 2. Детерминированность replay

Требования:

- порядок применения событий всегда определяется canonical `event_index`
- обработка nullable/unknown значений должна иметь стабильный fallback
- одинаковые входные данные всегда дают битово-эквивалентный (или семантически эквивалентный) результат

Запрещено:

- зависеть от недетерминированных источников времени/порядка
- silently пропускать события без диагностики

## 3. Валидация reconstructed state

Реконструкция проверяется против analyzer state:

- сравнение ключевых инвариантов (`turn`, `players`, `entities`, `resources`, `zones`)
- сравнение контрольных точек на границах ходов
- проверка, что materialized output из reconstructed state эквивалентен baseline output

## 4. Диагностика рассинхронизации

При mismatch фиксировать:

- `match_id`, `turn_id`, `event_id`
- ожидаемое и фактическое значение ключевых полей
- тип рассинхронизации (`missing_event`, `payload_conflict`, `order_conflict`, `unknown_resolution`)

Стратегия поведения:

- сохранить диагностический артефакт
- не повреждать canonical storage
- пометить матч как `reconstruction_degraded` для последующего расследования

## 5. Integration and invariant tests

1. `TestReconstructStateAtTurnBoundary`
   - корректность состояния на начале/конце хода.
2. `TestReconstructStateAtEventIndex`
   - корректность на произвольном событии внутри хода.
3. `TestReplayDeterminismSameInputSameState`
   - повторная реконструкция идентична.
4. `TestReconstructionMatchesAnalyzerState`
   - reconstructed state согласован с analyzer-пайплайном.
5. `TestReconstructionDiagnosticsOnMismatch`
   - при рассинхронизации формируется полный diagnostic report.
6. `TestReconstructionWithIncompleteMatch`
   - корректное поведение для incomplete матчей.

## 6. Критерий завершения 5.2

Шаг `5.2` считается завершенным, когда:

1. Описан алгоритм reconstruction из snapshot + events.
2. Зафиксированы требования к детерминированности replay.
3. Определены правила валидации против analyzer state.
4. Определена стратегия диагностики рассинхронизации.
5. Определены integration/invariant тесты для проверки алгоритма.
