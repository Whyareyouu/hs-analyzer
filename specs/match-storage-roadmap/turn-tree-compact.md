# Compact Turn Tree Design (Task 4.2)

## Цель

Определить навигационную структуру `match -> turns -> event refs`, где узел хода хранит метаданные и ссылки, а не дублирует полный state dump.

## 1. Навигационная модель

```text
Match
  -> Turn(1)
     -> EventRef[0..N]
  -> Turn(2)
     -> EventRef[0..N]
```

`EventRef` - ссылка на каноническое событие (`event_id`, `event_index`), а не inline-повтор payload.

## 2. Что хранится в узле `turn`

### Обязательные метаданные

- `turn_id`
- `match_id`
- `turn_number`
- `active_player_ref`
- `event_start_index`
- `event_end_index`
- `event_count`
- `started_at` / `ended_at` (nullable)

### Навигационные ссылки

- `event_refs[]` (упорядоченный массив/индекс ссылок)
- `snapshot_before_ref` (nullable)
- `snapshot_after_ref` (nullable)
- `decision_point_refs[]` (nullable)

## 3. Что не хранится в узле `turn`

Запрещено дублировать:

- полный state на каждый event
- полный `card info` на каждую ссылку
- историю всех entity-тегов внутри узла хода

Эти данные должны читаться по ссылкам из словарей/таблиц.

## 4. Быстрый доступ к одному ходу

Чтобы не читать весь матч:

1. По `match_id + turn_number` найти `turn_id`.
2. Получить `event_refs` по диапазону индексов.
3. При необходимости резолвить связанные сущности/карты лениво.

Целевой принцип:

- O(1) по индексу хода
- O(k) по числу событий выбранного хода

## 5. Вынос данных в отдельные структуры

Отдельно хранятся:

- канонические `event` записи
- entity dictionary матча
- card dictionary/refs
- snapshots/checkpoints
- decision points

## 6. Tests для навигации и выборки

1. `TestTurnLookupByMatchAndNumber`
   - точный поиск узла хода.
2. `TestTurnEventRefsRangeConsistency`
   - согласованность `event_start_index/event_end_index` и `event_count`.
3. `TestReadSingleTurnWithoutFullMatchLoad`
   - выборка одного хода без полной materialization матча.
4. `TestTurnNavigationStableAfterReindex`
   - сохранение навигационной целостности после rebuild/reindex.

## 7. Критерий завершения 4.2

Шаг `4.2` считается завершенным, когда:

1. Зафиксирована структура дерева `match -> turns -> event refs`.
2. Определены обязательные метаданные узла хода.
3. Явно зафиксировано, что не хранится в узле хода.
4. Определены тесты на навигацию и выборку одного хода.
