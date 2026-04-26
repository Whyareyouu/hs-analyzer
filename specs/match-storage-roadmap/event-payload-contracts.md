# Event Payload Contracts (Task 4.4)

## Цель

Зафиксировать минимальный payload событий для канонического event log, достаточный для реконструкции матча и materialization.

## 1. Общий контракт канонического события

Обязательные поля каждого event:

- `event_id`
- `match_id`
- `turn_id` (nullable)
- `event_index`
- `event_type`
- `occurred_at` (nullable)
- `source_log`
- `payload_version`
- `payload`

Где `payload` - типизированный JSON-объект по `event_type`.

## 2. Базовые event payload типы

## 2.1 `play`

Минимальный payload:

- `player_ref`
- `entity_ref` (nullable)
- `card_ref`
- `from_zone` (nullable)
- `to_zone`
- `resource_cost` (nullable)
- `source_ref` (nullable)

## 2.2 `draw`

Минимальный payload:

- `player_ref`
- `entity_ref` (nullable)
- `card_ref` (nullable до раскрытия)
- `from_zone`
- `to_zone`

## 2.3 `attack`

Минимальный payload:

- `attacker_entity_ref`
- `target_entity_ref` (nullable для face)
- `attacker_player_ref`
- `target_player_ref` (nullable)
- `attack_sequence_index` (nullable)

## 2.4 `damage`

Минимальный payload:

- `target_entity_ref` (nullable)
- `target_player_ref` (nullable)
- `source_entity_ref` (nullable)
- `amount`
- `damage_kind` (`combat`, `spell`, `effect`, `unknown`)

## 2.5 `heal`

Минимальный payload:

- `target_entity_ref` (nullable)
- `target_player_ref` (nullable)
- `source_entity_ref` (nullable)
- `amount`

## 2.6 `death`

Минимальный payload:

- `entity_ref`
- `zone_before`
- `zone_after` (обычно graveyard)
- `death_reason` (nullable)

## 2.7 `discover`

Минимальный payload:

- `player_ref`
- `source_entity_ref` (nullable)
- `options` (`card_ref[]` или `card_stub[]`)
- `chosen_card_ref` (nullable до фиксации выбора)

## 2.8 `effect_triggered`

Минимальный payload:

- `effect_source_entity_ref` (nullable)
- `effect_source_card_ref` (nullable)
- `trigger_context` (nullable)
- `target_refs` (nullable list)

## 2.9 `zone_change`

Минимальный payload:

- `entity_ref` (nullable)
- `card_ref` (nullable)
- `player_ref` (nullable)
- `from_zone`
- `to_zone`
- `reason` (nullable)

## 3. Timestamp, actor/target, resources, effect source

Общие правила:

- `occurred_at` хранится, если доступен из логов; иначе nullable.
- `actor/target` связи выражаются через refs, не через дубли объектов.
- Изменения ресурсов фиксируются в event payload, где это семантически уместно (`play`, `resource update` events).
- Источник эффекта (`effect_source_*`) обязателен, когда его можно извлечь детерминированно.

## 4. Backward-compatible расширение схемы

Разрешается:

- добавлять новые optional поля
- добавлять новые `event_type` с собственной `payload_version`
- вводить новые enum значения при сохранении fallback `unknown`

Запрещается без миграции:

- удалять обязательные поля существующего payload_version
- менять семантику существующего поля без version bump
- переиспользовать старый `event_type` для несовместимого payload

## 5. Unit/regression tests для payload contracts

1. `TestPlayPayloadContractV1`
2. `TestDrawPayloadContractV1`
3. `TestAttackPayloadContractV1`
4. `TestDamageHealDeathPayloadContractV1`
5. `TestDiscoverPayloadContractV1`
6. `TestEffectTriggeredPayloadContractV1`
7. `TestZoneChangePayloadContractV1`
8. `TestPayloadBackwardCompatibilityRules`
9. `TestUnknownEnumFallbackBehavior`
10. `TestEventPayloadRoundTripSerialization`

## 6. Критерий завершения 4.4

Шаг `4.4` считается завершенным, когда:

1. Определен минимальный payload для базовых типов событий.
2. Зафиксированы правила timestamp/refs/resources/effect-source.
3. Определены backward-compatible правила расширения event schema.
4. Определен набор unit/regression тестов на контракты payload.
