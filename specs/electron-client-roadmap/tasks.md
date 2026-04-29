# Implementation Plan

## Правило приемки

- Для каждого пункта и подпункта implementation plan должны быть определены и реализованы соответствующие тесты.
- Корректно выполненные тесты являются обязательной основой приемки задачи.
- Задача не считается завершенной, если UI собран, но отсутствуют или не проходят тесты, подтверждающие поведение.
- Для каждого этапа заранее определяется уровень проверки: unit, component, contract, integration, visual regression, end-to-end или performance test.

- [x] 1. Сформировать архитектурное решение для `hs-analyzer-client`

  - Зафиксировать, что клиент становится отдельным desktop-продуктом в монорепозитории
  - Зафиксировать разделение ответственности между `hs-analyzer-cli`, `hs-analyzer-client` и `hs-analyzer-be`
  - Зафиксировать, что клиент не парсит логи напрямую, а потребляет versioned local live contract
  - Подготовить architecture acceptance tests и dependency checks для Electron/FSD boundaries
  - _Requirements: 1, 6, 7, 10, 11_

  - [x] 1.1 Определить целевую структуру папок

    - Подготовить target layout `apps/hs-analyzer-client/electron` + `apps/hs-analyzer-client/src`
    - Зафиксировать FSD layout `app/pages/widgets/features/entities/shared`
    - Зафиксировать место для shared tokens, preload bridge и local API client
    - Подготовить static tests/checks на корректное направление зависимостей

  - [x] 1.2 Зафиксировать Electron security boundary

    - Зафиксировать `contextIsolation: true`
    - Зафиксировать `nodeIntegration: false`
    - Зафиксировать typed preload bridge как единственный способ доступа renderer к native API
    - Подготовить tests/checks на security configuration и preload API surface

- [x] 2. Спроектировать local live transport между CLI и клиентом

  - Определить `overlay_snapshot/v1`
  - Определить preferred realtime delivery channel
  - Зафиксировать initial snapshot + incremental updates
  - Зафиксировать degraded/reconnect semantics
  - Подготовить contract tests на valid/invalid payloads
  - _Requirements: 3, 4, 5, 6, 11_

  - [x] 2.1 Зафиксировать live payload модели

    - Определить `player_deck_remaining[]`
    - Определить `opponent_played_cards[]`
    - Зафиксировать, что это два разных массива и две разные view-model
    - Определить общие поля карточки: `name`, `description`, `image`, `mana_cost`
    - Подготовить contract tests на schema versioning и decoding

  - [x] 2.2 Подготовить CLI-side изменения под UI

    - Определить, как CLI вычисляет active deck
    - Определить, как CLI обновляет `remaining_copies` при draw/return-to-deck
    - Определить, как CLI агрегирует `opponent_played_cards`
    - Зафиксировать recovery behavior при mid-match restart
    - Подготовить integration tests на корректность локальных projections

- [x] 3. Подготовить foundation клиента на `Electron + React + Vite + Tailwind + shadcn/ui`

  - Создать приложение `apps/hs-analyzer-client`
  - Подключить Vite build для renderer
  - Подключить Tailwind
  - Подключить `shadcn/ui`
  - Подготовить smoke tests и build tests
  - _Requirements: 1, 9, 10, 11_

  - [x] 3.1 Подготовить design tokens

    - Вынести цвета, размеры, радиусы, shadows, motion, overlay dimensions в CSS variables
    - Связать Tailwind theme с design tokens
    - Подготовить tests на наличие и использование ключевых tokens

  - [x] 3.2 Подготовить app shell foundation

    - Добавить router/providers/theme
    - Подготовить shared layout primitives на `shadcn/ui`
    - Подготовить renderer smoke tests

- [x] 4. Реализовать overlay window runtime

  - Создать `OverlayWindow`
  - Реализовать позиционирование поверх Hearthstone
  - Реализовать lifecycle hide/show/recover
  - Подготовить integration tests на overlay lifecycle
  - _Requirements: 2, 2.1, 11_

  - [x] 4.1 Реализовать passive mode

    - Сделать overlay click-through по умолчанию
    - Зафиксировать, что он не блокирует взаимодействие с Hearthstone
    - Подготовить runtime tests на pointer behavior

  - [x] 4.2 Реализовать inspect mode

    - Добавить toggle/hotkey/menu action для интерактивного режима
    - Разрешить hover/focus interaction в inspect mode
    - Подготовить tests на переходы passive <-> inspect

  - [x] 4.3 Обработать ограничения window mode

    - Зафиксировать baseline support для windowed/borderless
    - Подготовить diagnostic state для unsupported fullscreen cases
    - Подготовить compatibility tests на supported/degraded modes

- [x] 5. Реализовать overlay UI для моей колоды

  - Построить левую верхнюю панель с полной активной колодой
  - Отобразить count каждой карты
  - Обновлять `remaining_copies` в live runtime
  - Подготовить component/e2e tests на draw-driven updates
  - _Requirements: 3, 5, 9, 11_

  - [x] 5.1 Реализовать row rendering

    - Отобразить название карты
    - Отобразить count
    - Использовать токенизированные размеры `~300px x ~20px`
    - Использовать background art strip
    - Подготовить visual regression tests для row layout

  - [x] 5.2 Реализовать zero-count behavior

    - Определить visual state для карт с `remaining_copies = 0`
    - Не удалять silently карточку из списка
    - Подготовить component tests на zero-count styling/state

- [x] 6. Реализовать overlay UI для карт оппонента

  - Построить правую панель opponent played cards
  - Отображать карты сразу после их розыгрыша
  - Агрегировать дубликаты по `play_count`
  - Подготовить component/e2e tests на update flow
  - _Requirements: 4, 5, 9, 11_

  - [x] 6.1 Реализовать sorted/aggregated presentation

    - Определить deterministic sorting для правого списка
    - Отобразить `play_count`
    - Подготовить tests на aggregation semantics

  - [x] 6.2 Реализовать reset semantics для нового матча

    - Очистить правую панель при старте новой игры
    - Не смешивать карты между матчами
    - Подготовить integration/regression tests на multi-game flow

- [x] 7. Реализовать tooltip/popover карточки

  - Добавить hover/focus reveal в inspect mode
  - Показывать изображение, название и описание
  - Реализовать fallback без картинки
  - Подготовить component/integration tests на tooltip behavior
  - _Requirements: 2.1, 5, 9, 11_

  - [x] 7.1 Подготовить metadata normalization

    - Зафиксировать формат очищенного текста карты
    - Зафиксировать fallback для отсутствующего описания или картинки
    - Подготовить tests на normalization/fallback

- [x] 8. Реализовать desktop dashboard

  - Построить стартовый экран со списком колод слева
  - Построить deck detail справа
  - Показать summary stats, matchup breakdown и список карт
  - Сохранить композицию, соответствующую `dashboard-visual-reference.md`
  - Подготовить page/widget integration tests
  - _Requirements: 7, 8, 9, 10, 11_

  - [x] 8.1 Реализовать deck sidebar

    - Отобразить все колоды пользователя
    - Отобразить краткую статистику рядом с каждой колодой
    - Реализовать выраженный selected state, соответствующий visual reference
    - Подготовить component tests на selection state

  - [x] 8.2 Реализовать deck detail summary

    - Показать wins/losses/winrate/total matches
    - Показать three-panel summary row без сторонней chart UI library
    - Реализовать крупный winrate ring как hero element summary блока
    - Подготовить visual regression и component tests

  - [x] 8.3 Реализовать deck cards grid

    - Показать карты выбранной колоды
    - Отобразить copies и служебные индикаторы
    - Сохранить dense multi-column deck-builder presentation из visual reference
    - Подготовить component tests и loading/empty state tests

- [x] 9. Интегрировать backend данные для dashboard

  - Подключить deck overview API
  - Подключить matchup API
  - Подключить deck cards API
  - Реализовать loading/error/offline/stale states
  - Подготовить contract и integration tests
  - _Requirements: 7, 8, 11_

  - [x] 9.1 Зафиксировать typed API client

    - Реализовать decoding/validation ответов
    - Подготовить tests на invalid payload handling

  - [x] 9.2 Реализовать cache/reconnect policy

    - Зафиксировать поведение при временной недоступности backend
    - Подготовить regression tests на stale/offline states

- [x] 10. Подготовить hardening, packaging и acceptance flow

  - Проверить long-running overlay stability
  - Проверить DPI/resolution compatibility
  - Подготовить production packaging
  - Подготовить e2e acceptance flow
  - _Requirements: 2, 7, 9, 11_

  - [x] 10.1 Подготовить performance and soak acceptance

    - Определить допустимую latency UI update после live event
    - Определить допустимый memory footprint overlay
    - Подготовить performance tests

  - [x] 10.2 Подготовить финальный end-to-end сценарий

    - Запустить CLI live runtime
    - Подключить Electron client
    - Получить live snapshot
    - Сыграть draw/play updates
    - Открыть dashboard и выбрать колоду
    - Подтвердить поведение e2e тестами
