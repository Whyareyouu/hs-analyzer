# Migration Strategy: `timeline_*.json` to Canonical Match Storage (Task 9)

## Цель

Определить безопасный переход от session-level `timeline_*.json` к canonical match-level storage без потери совместимости.

## 1. Какие артефакты сохраняются

Старые JSON-артефакты остаются как:

- `debug/export compatibility mode`
- временный канал верификации эквивалентности

Они больше не являются source of truth.

## 2. Миграционный путь

Этапы:

1. Включить dual-write:
   - canonical storage write
   - legacy timeline export
2. Включить compatibility checks:
   - сравнение materialized output из canonical storage с legacy output
3. Перевести downstream-потребителей на canonical materializer API.
4. Оставить legacy timeline только по флагу.

## 3. Проверка совпадения нового и старого вывода

Сравниваются:

- order/type ключевых событий
- turn segmentation
- ключевые факты действий/исходов

Допустимые расхождения:

- форматирование текста
- debug-only поля

## 4. Когда делать legacy timeline deprecated

Условия deprecation:

1. Пройден compatibility regression suite на репрезентативном наборе матчей.
2. Все критические потребители используют canonical materializer.
3. Нет открытых блокирующих расхождений по семантике output.

После этого:

- legacy режим доступен только как explicit compatibility flag

## 5. Migration/compatibility tests

1. `TestDualWriteCanonicalAndLegacy`
2. `TestCanonicalMaterializedEqualsLegacySemantics`
3. `TestMigrationNoDataLossAcrossMultipleMatches`
4. `TestLegacyCompatibilityFlagBehavior`
5. `TestDeprecatedLegacyPathWarning`

## 6. Критерий завершения 9

Шаг `9` считается завершенным, когда:

1. Описана стратегия dual-write и перехода на canonical storage.
2. Описана методика проверки эквивалентности output.
3. Определены условия deprecation legacy timeline.
4. Определены migration и compatibility regression tests.
