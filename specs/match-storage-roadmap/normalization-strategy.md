# Normalization Strategy for Card and Entity Data (Task 4.3)

## Цель

Уменьшить дублирование в canonical storage за счет нормализации карточных и сущностных данных, сохранив корректное восстановление истории.

## 1. Нормализация карточных данных

### 1.1 Card dictionary

Отдельный словарь карточных метаданных:

- `card_id` (PK)
- `name`
- `card_type`
- `cost`
- `attack` (nullable)
- `health` (nullable)
- `rarity` (nullable)
- `class` (nullable)
- `text_normalized` (nullable)

События хранят `card_ref` (`card_id`) вместо повторяющихся полных `card.info`.

### 1.2 Inline vs reference правило для card data

- Inline в событии:
  - только минимум для локального понимания (`card_id`, иногда `name` при необходимости explain/debug)
- По ссылке:
  - полные статические метаданные карты

## 2. Нормализация сущностей матча

### 2.1 Entity dictionary на матч

`entity_ref = match_id + entity_id`

Хранится:

- `entity_id`
- `first_seen_event_id`
- `current_card_id` (nullable)
- `owner_player_ref` (nullable)
- `zone` (latest known, nullable)
- `lifecycle_status` (`active`, `destroyed`, `unknown`)

### 2.2 История привязки `entity_id -> card_id`

Отдельный журнал соответствий:

- `entity_ref`
- `event_id`
- `card_id`
- `binding_reason` (`full_entity`, `show_entity`, `tag_change`, `generated`)

Это позволяет корректно обработать случаи смены/уточнения card identity.

## 3. Generated cards и скрытая информация

### 3.1 Generated cards

- Generated card должна иметь связь с событием генерации.
- До раскрытия card identity допускается placeholder-сущность с `card_id = unknown`.
- После раскрытия обновление делается через binding history, без перезаписи прошлых событий.

### 3.2 Hidden/uncertain data

- Данные, недоступные игроку в момент решения, помечаются `visibility_scope`.
- Для decision analysis должны быть доступны:
  - `public_view`
  - `player_view` (опционально по игроку)
- Запрещено silently смешивать hidden и public признаки в одном аналитическом представлении.

## 4. Что хранится inline в событии

Inline сохраняются:

- refs (`entity_ref`, `card_ref`, `actor_ref`, `target_ref`)
- динамические числовые значения события (урон, хил, ресурсы, zone transition)
- event-local флаги

По ссылкам читаются:

- статические карточные метаданные
- накопленная сущностная история

## 5. Тесты на дедупликацию и корректность ссылок

1. `TestCardInfoDedupAcrossEvents`
   - одно и то же `card_id` не дублируется полным объектом по событиям.
2. `TestEntityCardBindingHistoryResolution`
   - корректное восстановление `entity -> card` на любом event index.
3. `TestGeneratedCardResolution`
   - generated card корректно проходит путь `unknown -> known`.
4. `TestHiddenDataSeparation`
   - hidden/uncertain данные не попадают в public decision context.
5. `TestEventReferenceIntegrity`
   - все refs в событиях резолвятся в существующие dictionary записи.

## 6. Критерий завершения 4.3

Шаг `4.3` считается завершенным, когда:

1. Определены card/entity dictionaries и правила ссылок.
2. Зафиксировано, что хранится inline, а что по ссылке.
3. Зафиксирована стратегия для generated/hidden data.
4. Определен набор тестов на дедупликацию и резолв ссылок.
