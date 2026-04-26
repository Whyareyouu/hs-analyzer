# Timeline Compatibility Reconstruction (Task 6.2)

## Цель

Определить, как из нового canonical storage восстанавливается эквивалент текущих `timeline_*` представлений (`human` и `replay`) без потери ключевой семантики.

## 1. Реконструкция `human timeline`

Из canonical данных формируются:

- матчевые метаданные
- список ходов
- человекочитаемые события с резолвом card/entity refs
- агрегаты по ходу (рука, стол, ключевые эффекты)

Требование:

- итоговое представление должно оставаться читаемым для человека без технического шума.

## 2. Реконструкция `replay timeline`

Из canonical данных формируются:

- детализированная последовательность событий по `event_index`
- структуры `opening_state`/`state_after` (или их эквивалент)
- refs на сущности/карты и причины переходов

Требование:

- replay-вид должен быть пригоден для детерминированного анализа и debug.

## 3. Optional/debug-only поля

Допускается перенос части полей в optional/debug:

- сырые parser-specific промежуточные атрибуты
- heavy diagnostic blobs
- низкоуровневые trace поля

Правило:

- ключевые факты матча (действия, эффекты, исход) не могут становиться debug-only.

## 4. Критерии эквивалентности старого и нового output

Эквивалентность считается выполненной, если:

1. Совпадает порядок и типы ключевых событий.
2. Совпадают ключевые факты (`play`, `draw`, `attack`, `damage`, `death`, `discover`, `zone_change`).
3. Совпадает turn-level структура.
4. Расхождения допускаются только в:
   - форматировании текста
   - optional/debug полях

## 5. Regression tests

1. `TestHumanTimelineCompatibilityRegression`
2. `TestReplayTimelineCompatibilityRegression`
3. `TestTimelineEventOrderEquivalence`
4. `TestTimelineKeyFactsEquivalence`
5. `TestOptionalFieldsDoNotAffectSemanticEquivalence`

## 6. Критерий завершения 6.2

Шаг `6.2` считается завершенным, когда:

1. Описана реконструкция `human` и `replay` output из canonical storage.
2. Зафиксировано, какие поля допускаются как optional/debug-only.
3. Определены критерии эквивалентности старого и нового timeline output.
4. Определены regression tests на compatibility.
