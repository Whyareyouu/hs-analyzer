# Design Document: Electron Client

## Overview

Новая клиентская подсистема добавляет третий продукт в монорепозиторий:

- `hs-analyzer-cli` - локальный runtime для логов, live-monitoring и live projections
- `hs-analyzer-be` - серверный storage/history/analytics слой
- `hs-analyzer-client` - desktop UI на Electron + React для overlay и dashboard

Ключевое архитектурное правило:

**`hs-analyzer-client` не парсит Hearthstone логи напрямую.**  
Все live-данные о текущем матче приходят из `hs-analyzer-cli` через локальный versioned transport.

## Product Scope

Подсистема `electron-client-roadmap` покрывает:

- desktop shell window
- overlay window поверх Hearthstone
- live deck tracker слева сверху
- live opponent played cards справа
- tooltip/popover по картам
- dashboard со списком колод и статистикой
- FSD-архитектуру renderer-слоя
- Electron security boundary и IPC boundary

Подсистема не покрывает в v1:

- собственный парсинг логов внутри Electron клиента
- server-side spectator mode
- mobile/web client
- внешние UI kit кроме `shadcn/ui`
- отдельную chart library

## Target Architecture

```text
Hearthstone Logs
      |
      v
apps/hs-analyzer-cli
  log discovery
  live reader
  parsers
  session detector
  active deck projection
  opponent played projection
  local live transport
      |
      | localhost WS / HTTP
      v
apps/hs-analyzer-client
  Electron main
  preload bridge
  React renderer
    - overlay UI
    - dashboard UI
      |
      | HTTP
      v
apps/hs-analyzer-be
  deck history
  matchup analytics
  deck detail
  future match viewer
```

## Integration With Existing Roadmaps

`specs/hearthstone-log-analyzer` остается базой для parser/domain модели и card metadata enrichment.

`specs/live-log-monitoring-roadmap` остается базой для live runtime, session detector, local observability и live event stream.

`specs/match-storage-roadmap` остается базой для server-side match history, materialized views и analytics datasets, которые потом использует dashboard.

Новая клиентская подсистема не заменяет эти roadmap-ы, а добавляет над ними UI слой и локальный desktop runtime.

## Window Model

Клиент использует как минимум два окна:

1. `MainWindow`
   - обычное desktop-окно приложения
   - показывает список колод, статистику и аналитику

2. `OverlayWindow`
   - прозрачное frameless окно
   - показывается поверх Hearthstone
   - живет независимо от `MainWindow`

`OverlayWindow` не должен быть просто отдельной страницей внутри `MainWindow`. Это другой runtime surface с другими требованиями:

- прозрачность
- always-on-top
- позиционирование относительно окна игры
- click-through by default
- минимальная латентность обновления

## Overlay Interaction Constraint

Для overlay есть жесткое техническое ограничение: если окно полностью click-through, то normal hover/tooltip не работает. Поэтому v1 design фиксирует двухрежимную модель:

- `passive mode`
  - overlay визуально видим
  - мышь проходит сквозь него в Hearthstone
  - режим дефолтный во время игры

- `inspect mode`
  - overlay временно принимает pointer events
  - пользователь может наводить курсор на строки карт
  - tooltip/popover становится доступным

Без этой модели требование “hover показывает подробную карточку” будет конфликтовать с требованием “overlay не мешает игре”.

## Supported Hearthstone Window Modes

Поддерживаемый baseline для overlay:

- windowed
- borderless-windowed

`exclusive fullscreen` не должен считаться silently supported. Если на конкретной конфигурации overlay не может быть надежно показан поверх такого режима, клиент должен явно сообщать об этом пользователю.

## Live Data Model

Overlay использует две разные live-модели:

### 1. Player deck remaining model

Нужна для левой панели.

Примерные поля:

- `deck_id`
- `deck_name`
- `cards[]`
  - `card_id`
  - `name`
  - `mana_cost`
  - `initial_copies`
  - `remaining_copies`
  - `drawn_copies`
  - `image`
  - `description`

Это не тот же самый массив, что список карт оппонента.

### 2. Opponent played cards model

Нужна для правой панели.

Примерные поля:

- `cards[]`
  - `card_id`
  - `name`
  - `mana_cost`
  - `play_count`
  - `first_played_at`
  - `last_played_at`
  - `image`
  - `description`

Эта модель агрегирует уже сыгранные карты оппонента и не пытается угадывать его скрытую руку или колоду.

## UI Composition

### Overlay

Композиция overlay:

- слева сверху: `PlayerDeckPanel`
- справа: `OpponentPlayedPanel`
- tooltip layer: `CardDetailsPopover`

### Main Shell

Композиция основного окна:

- слева: `DeckSidebar`
- справа:
  - `DeckHeader`
  - `DeckSummaryCards`
  - `MatchupBreakdown`
  - `DeckCardsGrid`

Визуальный baseline опирается на приложенные референсы:

- dark premium background
- glass-like panels
- class-colored accents
- card art strips в списках

## CLI Changes Required For UI

Чтобы UI был корректным, потребуется расширение локального runtime `hs-analyzer-cli`:

1. Определять активную колоду игрока для текущего матча.
2. Строить `player_deck_remaining` projection.
3. Строить `opponent_played_cards` projection.
4. Резолвить metadata карты: имя, текст, изображение.
5. Отдавать live snapshot в отдельном versioned local contract.
6. Сохранять recovery semantics при рестарте monitor/runtime.

Это допустимая часть roadmap и не является нарушением границы между клиентом и CLI.

## Backend Dependencies For Dashboard

Overlay зависит от CLI, а dashboard зависит от backend.

Это intentional split:

- live overlay должен продолжать работать даже при недоступном backend
- history/statistics dashboard может показывать `loading`, `error`, `stale cache` или `offline` состояния

Таким образом:

- live runtime availability != backend availability
- overlay uptime не должен ломаться из-за server-side analytics

## Renderer Architecture

Renderer использует FSD:

```text
src/
  app/
  pages/
  widgets/
  features/
  entities/
  shared/
```

Ключевые правила:

- `pages` собирают сценарий целого экрана
- `widgets` собирают крупные смысловые блоки
- `features` инкапсулируют интеракции и use-cases
- `entities` описывают deck/card/match/live-overlay сущности
- `shared` содержит UI primitives, tokens, helpers, API client и typed contracts

Electron `main` и `preload` код живут отдельно от renderer FSD-слоев.

## Styling Rules

Tailwind используется как utility layer поверх design tokens.

Главное правило:

**никаких раскиданных по коду hardcoded цветов, размеров, радиусов и overlay offsets.**

Все ключевые величины выносятся в CSS variables:

- colors
- typography sizes
- spacing
- radii
- shadows
- z-index
- animation durations
- overlay row width/height
- overlay panel offsets

## Completion Criteria

Подсистема считается принятой только когда:

1. есть отдельное приложение `hs-analyzer-client`
2. overlay корректно показывает left deck tracker и right opponent played cards
3. card tooltip работает в inspect mode
4. main dashboard показывает список колод и детальную статистику
5. клиент соблюдает Electron security boundary
6. все этапы подтверждены тестами
