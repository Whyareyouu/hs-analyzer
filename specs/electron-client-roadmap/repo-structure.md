# Repo Structure

## Purpose

Зафиксировать целевую структуру папок для добавления desktop-клиента в текущий монорепозиторий и определить границы между Electron main/preload/runtime кодом и FSD renderer-слоем.

## Target Layout

```text
apps/
  hs-analyzer-cli/
  hs-analyzer-be/
  hs-analyzer-client/
    electron/
      main/
      preload/
      shared/
    public/
    src/
      app/
        providers/
        router/
        styles/
      pages/
        deck-overview/
        live-overlay-control/
        settings/
      widgets/
        deck-sidebar/
        deck-summary/
        matchup-breakdown/
        deck-cards-grid/
        overlay-player-deck/
        overlay-opponent-cards/
      features/
        select-deck/
        toggle-overlay/
        inspect-card/
        reconnect-live-transport/
      entities/
        deck/
        card/
        live-match/
        matchup/
      shared/
        api/
        config/
        lib/
        model/
        styles/
        ui/
        assets/

specs/
  electron-client-roadmap/
```

## Folder Responsibilities

### `apps/hs-analyzer-client/electron/main`

Отвечает за:

- создание окон
- lifecycle Electron app
- позиционирование overlay
- native integration
- app menus / tray / shortcuts

### `apps/hs-analyzer-client/electron/preload`

Отвечает за:

- typed bridge между main и renderer
- безопасный export минимально необходимых API
- отсутствие утечки произвольных Node primitives в renderer

### `apps/hs-analyzer-client/src`

Отвечает за весь renderer UI на FSD.

## FSD Rules

### `app`

Глобальные providers:

- theme
- router
- query/cache provider
- overlay/live transport bootstrap

### `pages`

Полные экраны:

- main dashboard
- settings
- overlay control / diagnostics page

### `widgets`

Крупные композиционные блоки:

- sidebar со списком колод
- summary/stat cards
- matchup panel
- cards grid
- overlay panels

### `features`

Интерактивные use-cases:

- выбор колоды
- переключение overlay режима
- inspect карточки
- reconnect live transport

### `entities`

Предметные модели:

- `deck`
- `card`
- `live-match`
- `matchup`

### `shared`

Только переиспользуемые примитивы:

- shadcn/ui wrappers
- typed API client
- design tokens
- date/format helpers
- icons/assets mapping

`shared` не должен превращаться в свалку бизнес-логики.

## Monorepo Ownership Rules

### `hs-analyzer-cli`

Владеет:

- локальными live projections
- active match runtime
- local live transport contract

### `hs-analyzer-client`

Владеет:

- Electron shell
- overlay rendering
- dashboard rendering
- interaction logic

### `hs-analyzer-be`

Владеет:

- decks history
- match history
- matchup analytics
- future deck insights

## Cross-Boundary Rules

1. Renderer не импортирует код из Go runtime напрямую.
2. Electron renderer не читает Hearthstone logs с файловой системы напрямую.
3. Overlay данные приходят из local transport, а не из ad-hoc парсинга JSON-файлов без versioned contract.
4. Dashboard данные приходят из backend API или локального typed cache слоя.
5. `shadcn/ui` является единственным UI component library.

## Acceptance Tests

Required checks:

1. Static architecture tests/checks для FSD dependency direction
2. Electron security configuration tests
3. Renderer import boundary tests
4. Contract tests for preload API surface
5. UI test suite for critical widgets/pages
