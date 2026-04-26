# Source of Truth Strategy (Task 2.3)

## 1. Принцип истины

Primary storage содержит канонические факты матча. Все производные представления (`human log`, `compatibility JSON`, `replay/debug JSON`, graph projection) восстанавливаются из него.

Канонически хранятся:

- `match`
- `turn`
- `event`
- (расширения roadmap) `snapshot`, `decision_point`, `candidate_action`

## 2. Версионирование схемы и миграции

### 2.1 Версии

- Вводится поле `schema_version` на уровне базы и контрактов чтения/записи.
- Стартовая версия: `v1`.
- Любое изменение структуры таблиц/индексов фиксируется как новая миграция.

### 2.2 Правила миграций

- Миграции выполняются последовательно и атомарно.
- Каждая миграция должна быть обратимо проверяема через regression-контур.
- Данные старых матчей после миграции должны оставаться читаемыми без потери идентификаторов.

## 3. Формат идентификаторов

Идентификаторы задаются детерминированно, чтобы обеспечить стабильные ссылки и idempotent запись.

## 3.1 `match_id`

Формат:

`match:<game_id_or_fallback>`

Правила:

- если из событий восстановлен `game_id`, он используется как основа
- если `game_id` отсутствует, используется fallback-ключ на основе временного диапазона и hash последовательности событий

## 3.2 `turn_id`

Формат:

`turn:<match_id>:<turn_number>`

Пример:

`turn:match:123456:7`

## 3.3 `event_id`

Формат:

`event:<match_id>:<event_index>`

Где `event_index` - стабильный индекс события в каноническом порядке матча.

## 3.4 `snapshot_id`

Формат:

`snapshot:<match_id>:<anchor_type>:<anchor_value>`

Где:

- `anchor_type` in (`turn`, `decision_point`, `checkpoint`)
- `anchor_value` - номер хода или локальный индекс decision/checkpoint

## 3.5 `decision_point_id`

Формат:

`dp:<match_id>:<turn_number>:<dp_index>`

## 4. Правила стабильности ID

- Формат ID не меняется внутри одной major schema version.
- При миграциях не допускается silent-перегенерация ID для уже записанных матчей.
- Любое изменение формата ID требует явной migration strategy и compatibility tests.

## 5. Regression tests для 2.3

Обязательный набор:

1. `TestMigrationV1ToV2PreservesIDs`
   - проверяет, что `match_id/turn_id/event_id` не изменились после миграции.
2. `TestDeterministicIDGeneration`
   - проверяет, что повторная запись одного матча дает те же ID.
3. `TestFallbackMatchIDGeneration`
   - проверяет стабильность fallback `match_id`, если нет `game_id`.
4. `TestReadOldDataAfterMigration`
   - проверяет чтение матчей старой версии после применения новых миграций.
5. `TestProjectionRebuildKeepsReferences`
   - проверяет, что rebuild secondary projection сохраняет ссылки на canonical IDs.

## 6. Критерий завершения 2.3

Шаг `2.3` считается завершенным, когда:

1. Зафиксирована политика source of truth и разделения canonical/derived.
2. Определены правила schema versioning и миграций.
3. Определен формат `match_id`, `turn_id`, `event_id`, `snapshot_id`, `decision_point_id`.
4. Определен regression-набор тестов на стабильность ID и миграции.
