# Minimal Graph Schema (Task 3.3)

## Цель

Определить минимальную graph schema, достаточную для:

- query сценариев из `graph-use-cases.md`
- explainability цепочек `state -> action -> outcome`
- rebuild проекции из canonical storage

## 1. Узлы

Обязательные node types:

- `Match`
- `Turn`
- `DecisionPoint`
- `Snapshot`
- `Action`
- `Outcome`
- `Entity`
- `Card`

## 2. Минимальные свойства узлов

### Match

- `match_id` (PK reference)
- `schema_version`
- `game_mode`
- `started_at`
- `ended_at`

### Turn

- `turn_id`
- `match_id`
- `turn_number`
- `active_player`

### DecisionPoint

- `decision_point_id`
- `match_id`
- `turn_id`
- `phase` (`start_turn`, `post_draw`, `discover`, `combat_sequence`, `end_turn`)

### Snapshot

- `snapshot_id`
- `match_id`
- `turn_id`
- `visibility_scope` (`public`, `player_view`)

### Action

- `action_id` (derived from canonical event)
- `match_id`
- `turn_id`
- `action_type`
- `actor_ref`

### Outcome

- `outcome_id` (derived from canonical event or aggregation window)
- `match_id`
- `turn_id`
- `outcome_type`

### Entity

- `entity_ref` (`match_id + entity_id`)
- `match_id`
- `entity_id`
- `latest_card_id` (nullable)

### Card

- `card_id`
- `name`
- `card_type`
- `cost` (nullable)

## 3. Связи (edges)

Базовые связи:

- `Match -HAS_TURN-> Turn`
- `Turn -HAS_DECISION_POINT-> DecisionPoint`
- `DecisionPoint -HAS_SNAPSHOT_BEFORE-> Snapshot`
- `DecisionPoint -HAS_ACTION-> Action`
- `Action -LEADS_TO-> Outcome`
- `Outcome -UPDATES_TO-> Snapshot` (snapshot after)
- `Action -USES_ENTITY-> Entity`
- `Outcome -AFFECTS_ENTITY-> Entity`
- `Entity -REPRESENTS_CARD-> Card`
- `Turn -CONTAINS_ACTION-> Action`
- `Turn -CONTAINS_OUTCOME-> Outcome`

## 4. Правила построения из canonical storage

1. Источник данных - только canonical tables/records.
2. Все graph IDs должны быть либо canonical IDs, либо детерминированно derived IDs.
3. Запрещено создавать graph-only факты, которые нельзя восстановить из canonical storage.
4. При конфликте приоритета данных canonical storage всегда выигрывает.

## 5. Правила переиндексации (reindex/rebuild)

### 5.1 Full rebuild

- Удалить текущую projection схему.
- Пройти все матчи в canonical storage по стабильному порядку.
- Пересобрать nodes/edges детерминированно.

### 5.2 Incremental reindex

- Для новых/обновленных матчей пересобрать только соответствующие subgraph.
- Удалить stale edges перед повторной вставкой.

### 5.3 Versioned projection

- Projection version хранится отдельно от canonical schema version.
- Смена projection version допускает полную перестройку без изменения canonical данных.

## 6. Schema validation tests

1. `TestGraphSchemaNodeTypesPresent`
   - проверяет наличие обязательных типов узлов.
2. `TestGraphSchemaEdgeTypesPresent`
   - проверяет наличие обязательных типов ребер.
3. `TestGraphNodesReferenceCanonicalIDs`
   - проверяет ссылки на существующие canonical IDs.
4. `TestNoOrphanDecisionChain`
   - проверяет целостность цепочки `DecisionPoint -> Action -> Outcome`.
5. `TestCardEntityLinkResolution`
   - проверяет корректность `Entity -> Card` связей, включая generated cards.

## 7. Query regression tests

1. `TestQuerySimilarDecisionPointsRegression`
2. `TestQueryStateActionOutcomePathRegression`
3. `TestQueryExplainabilityTraceRegression`
4. `TestQueryEffectChainRegression`

Каждый тест фиксирует expected shape ответа и стабильность ссылок на canonical IDs.

## 8. Критерий завершения 3.3

Шаг `3.3` считается завершенным, когда:

1. Определены обязательные узлы и связи минимальной графовой схемы.
2. Зафиксирован минимальный набор свойств для аналитики и explainability.
3. Описаны правила reindex/rebuild из primary storage.
4. Определены schema validation и query regression tests.
