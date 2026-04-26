# Canonical Match Structure (Task 4.1)

## Цель

Определить каноническую структуру хранения одного матча, совместимую с event-sourced подходом и будущими расширениями (`snapshot`, `decision_point`, `candidate_action`).

## 1. Обязательные сущности v1

Минимальный обязательный набор для первого runtime-этапа:

- `match`
- `turn`
- `event`

Расширяемые сущности:

- `entity`
- `snapshot`
- `decision_point`
- `candidate_action`

## 2. Границы матча и завершение

### 2.1 Границы

- Матч начинается с первого `GameStart`-совместимого события, относящегося к новой игровой сессии.
- Матч завершается на `GameEnd` или по fallback-правилам finalization (если строгий `GameEnd` отсутствует).

### 2.2 Неполные матчи

Неполный матч может быть записан при:

- crash/restart мониторинга
- закрытии приложения во время активной игры
- поврежденном или частичном логе

Такая запись должна иметь:

- `completion_status = incomplete`
- `incomplete_reason`
- сохраненный event stream без silent-drop событий

## 3. Контракт сущности `match`

### Обязательные поля

- `match_id`
- `schema_version`
- `source_session_id` (nullable)
- `game_id` (nullable, если недоступен в логах)
- `game_mode`
- `started_at`
- `ended_at` (nullable для incomplete)
- `completion_status` (`complete` | `incomplete`)
- `completion_reason`
- `players_summary` (минимальные метаданные)
- `turn_count`
- `event_count`
- `created_at`
- `updated_at`

### Опциональные поля

- `winner_player_ref`
- `deck_refs`
- `tags` (debug/diagnostics)

## 4. Контракт сущности `turn`

### Обязательные поля

- `turn_id`
- `match_id`
- `turn_number`
- `active_player_ref`
- `started_at` (nullable)
- `ended_at` (nullable)
- `event_start_index`
- `event_end_index`
- `event_count`

### Опциональные поля

- `opening_state_ref`
- `closing_state_ref`
- `decision_point_count`

## 5. Контракт сущности `event`

### Обязательные поля

- `event_id`
- `match_id`
- `turn_id` (nullable для pre-turn/prelude событий)
- `event_index` (глобально в матче)
- `event_type`
- `occurred_at` (nullable если нет timestamp в источнике)
- `source_log`
- `payload` (versioned JSON/BLOB)

### Опциональные поля

- `actor_ref`
- `target_ref`
- `entity_ref`
- `card_ref`
- `debug_context`

## 6. Schema versioning

- Вводится `schema_version` на уровне storage schema и на уровне match contract.
- Начальная версия: `match_schema_v1`.
- Любое изменение обязательных полей или формата payload требует migration path и regression tests.

## 7. Schema contract tests

Обязательные тесты:

1. `TestWriteReadMatchContractV1`
   - запись и чтение полного `match` с обязательными полями.
2. `TestWriteReadTurnContractV1`
   - запись/чтение `turn` и корректных event range ссылок.
3. `TestWriteReadEventContractV1`
   - запись/чтение `event` с payload без потери данных.
4. `TestIncompleteMatchContract`
   - проверка сохранения incomplete статуса и причины.
5. `TestSchemaVersionCompatibility`
   - проверка поведения при чтении поддерживаемых/неподдерживаемых версий.

## 8. Критерий завершения 4.1

Шаг `4.1` считается завершенным, когда:

1. Формально определены сущности и поля canonical match structure.
2. Зафиксированы правила границ/завершения/неполных матчей.
3. Зафиксирован schema versioning контракт.
4. Определен обязательный набор schema contract tests.
