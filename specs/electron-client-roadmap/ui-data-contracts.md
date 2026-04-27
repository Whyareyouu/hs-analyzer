# UI Data Contracts

## Purpose

Определить data contracts между:

- `hs-analyzer-cli` -> `hs-analyzer-client` для live overlay
- `hs-analyzer-be` -> `hs-analyzer-client` для dashboard/history

## 1. Local Live Transport

### Goal

`hs-analyzer-client` должен получать live данные о текущем матче push-способом. Предпочтительный транспорт для v1:

- `GET /api/v1/live/overlay-snapshot`
- `WS /ws/live`

Файловый snapshot может оставаться fallback/recovery adapter, но не primary realtime API.

### Overlay Snapshot Contract

Версия:

- `overlay_snapshot/v1`

Пример payload:

```json
{
  "schema_version": "overlay_snapshot/v1",
  "match": {
    "match_id": "match-123",
    "game_id": "game-123",
    "status": "active",
    "player_class": "DEMONHUNTER",
    "opponent_class": "MAGE",
    "deck_id": "deck-spell-dh",
    "deck_name": "Spell DH"
  },
  "player_deck_remaining": [
    {
      "card_id": "CARD_001",
      "name": "Card Name",
      "mana_cost": 2,
      "initial_copies": 2,
      "remaining_copies": 1,
      "drawn_copies": 1,
      "image": "https://...",
      "description": "Normalized card text"
    }
  ],
  "opponent_played_cards": [
    {
      "card_id": "CARD_777",
      "name": "Opponent Card",
      "mana_cost": 3,
      "play_count": 2,
      "first_played_at": "2026-04-27T12:00:00Z",
      "last_played_at": "2026-04-27T12:05:00Z",
      "image": "https://...",
      "description": "Normalized card text"
    }
  ],
  "health": {
    "monitor": "healthy",
    "transport": "healthy"
  }
}
```

### Contract Rules

1. `player_deck_remaining[]` и `opponent_played_cards[]` - это разные массивы и разные UI модели.
2. `player_deck_remaining[]` сортируется детерминированно.
3. `opponent_played_cards[]` агрегирует plays по `card_id`.
4. Payload должен содержать нормализованные поля, пригодные для непосредственного UI-рендера.
5. Изменение схемы требует явного version bump.

### Live Events Over WebSocket

Поддерживаемые сообщения:

- `overlay_snapshot`
- `overlay_delta`
- `overlay_reset`
- `transport_health`

`overlay_delta` допускает patch-like обновления, но клиент всегда должен уметь восстановиться из полного `overlay_snapshot`.

## 2. CLI Changes Required For Overlay Contract

CLI должен дополнительно уметь:

1. Определять активную колоду игрока.
2. Поддерживать счетчики `initial_copies`, `remaining_copies`, `drawn_copies`.
3. Поддерживать отдельный projection для карт, разыгранных оппонентом.
4. Отдавать live payload через local server/transport.
5. Восстанавливать live snapshot после mid-match restart без silent corruption.

## 3. Backend Dashboard Contracts

### `GET /api/v1/decks`

Возвращает список колод пользователя.

Пример полей:

- `deck_id`
- `name`
- `class`
- `wins`
- `losses`
- `winrate`
- `last_played_at`
- `icon`

### `GET /api/v1/decks/:deckId/overview`

Возвращает summary для заголовка и overview cards:

- `deck_id`
- `name`
- `class`
- `wins`
- `losses`
- `total_matches`
- `winrate`
- `best_matchups[]`
- `worst_matchups[]`

### `GET /api/v1/decks/:deckId/cards`

Возвращает список карт колоды:

- `card_id`
- `name`
- `mana_cost`
- `copies`
- `rarity`
- `is_legendary`
- `image`

### `GET /api/v1/decks/:deckId/matchups`

Возвращает analytics dataset для matchup panel:

- `opponent_class` или `opponent_archetype`
- `wins`
- `losses`
- `winrate`
- `matches`

## 4. Client Caching Rules

1. Overlay live data не кэшируется как long-term source of truth.
2. Dashboard data может кэшироваться краткоживуще на клиенте.
3. При недоступном backend UI может показывать stale cache, но это должно быть явно обозначено.
4. Local transport reconnect не должен требовать перезапуска desktop app.

## 5. Acceptance Tests

Required tests:

1. Contract tests для `overlay_snapshot/v1`
2. Contract tests для WebSocket message envelope
3. CLI integration tests на корректное разделение `player_deck_remaining[]` и `opponent_played_cards[]`
4. Backend API contract tests для deck overview/matchups/cards
5. Client decoding tests для valid/invalid payloads
