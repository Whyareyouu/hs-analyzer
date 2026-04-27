# Requirements Document: Electron Client

## Introduction

Подсистема `electron-client-roadmap` определяет создание отдельного desktop-клиента `hs-analyzer-client`, который отвечает за:

- overlay поверх Hearthstone в live-режиме
- desktop shell на Electron + React + Vite
- отображение локальной live-информации из `hs-analyzer-cli`
- отображение deck history и аналитики из backend

Текущий `hs-analyzer-cli` остается владельцем парсинга логов, live-monitoring и локальной игровой телеметрии. Новый `hs-analyzer-client` не парсит логи напрямую и не дублирует parser/runtime-логику CLI.

Корректно выполненные тесты являются обязательной основой приемки для всех требований этой подсистемы.

## Requirements

### Requirement 1: Desktop client foundation

**User Story:** Как пользователь, я хочу отдельное desktop-приложение, чтобы видеть overlay и статистику не через CLI, а через полноценный UI.

#### Acceptance Criteria

1. WHEN реализуется desktop-клиент THEN он SHALL быть отдельным приложением `hs-analyzer-client`
2. WHEN создается клиент THEN стек renderer SHALL быть `React + Vite + Tailwind CSS + shadcn/ui`
3. WHEN создается desktop shell THEN он SHALL использовать Electron
4. WHEN renderer работает THEN он SHALL не иметь прямого небезопасного доступа к Node.js API
5. WHEN Electron конфигурируется THEN `contextIsolation` SHALL быть включен, `nodeIntegration` SHALL быть выключен, а доступ к native возможностям SHALL идти через typed preload bridge

### Requirement 2: Overlay runtime over Hearthstone

**User Story:** Как игрок, я хочу видеть overlay поверх самой игры, чтобы получать live-информацию, не переключаясь из окна Hearthstone.

#### Acceptance Criteria

1. WHEN активен live match THEN клиент SHALL показывать отдельное overlay window поверх Hearthstone
2. WHEN положение или размер окна Hearthstone меняются THEN overlay SHALL корректно следовать за ними
3. WHEN Hearthstone работает в поддерживаемом режиме THEN overlay SHALL поддерживать как минимум windowed и borderless-windowed режимы
4. IF Hearthstone работает в неподдерживаемом fullscreen режиме THEN клиент SHALL явно сообщать о деградации overlay, а не молча не показывать данные
5. WHEN overlay включен THEN он SHALL оставаться поверх игры без визуального захвата фокуса desktop shell окна

### Requirement 2.1: Overlay interaction model

**User Story:** Как пользователь, я хочу, чтобы overlay не мешал играть, но при необходимости позволял получить расширенную информацию по картам.

#### Acceptance Criteria

1. WHEN overlay показывается по умолчанию THEN он SHALL работать в click-through/passive режиме
2. WHEN пользователю нужна интерактивность THEN клиент SHALL иметь явный inspect mode или equivalent toggle для включения pointer events
3. WHEN overlay находится в passive режиме THEN он SHALL не блокировать взаимодействие с окном Hearthstone
4. WHEN overlay находится в inspect mode THEN пользователь SHALL иметь возможность навести курсор на карточку и открыть tooltip/popover
5. WHEN inspect mode выключается THEN overlay SHALL возвращаться в click-through режим

### Requirement 3: Player deck tracker panel

**User Story:** Как игрок, я хочу видеть мою колоду слева сверху со всеми картами и количеством копий, чтобы отслеживать remaining deck в реальном времени.

#### Acceptance Criteria

1. WHEN матч начинается и активная колода определена THEN overlay SHALL показать список всех карт моей колоды в левом верхнем блоке
2. WHEN карта присутствует в стартовой колоде THEN система SHALL отображать ее с количеством копий
3. WHEN я беру карту из колоды THEN remaining count соответствующей карты SHALL уменьшаться немедленно
4. WHEN карта возвращается обратно в колоду valid game-effect сценарием THEN remaining count SHALL увеличиваться обратно
5. WHEN карта не входила в исходную колоду и была сгенерирована эффектом THEN она SHALL не попадать в список remaining deck как карта основной колоды
6. WHEN remaining count карты становится равным нулю THEN UI SHALL визуально отражать, что все копии уже вышли из колоды

### Requirement 4: Opponent played cards panel

**User Story:** Как игрок, я хочу видеть справа список всех карт, которые уже разыграл оппонент, чтобы быстрее понимать его линию игры.

#### Acceptance Criteria

1. WHEN оппонент разыгрывает карту THEN она SHALL появляться в правом списке немедленно
2. WHEN оппонент повторно разыгрывает ту же карту THEN UI SHALL обновлять счетчик play count, а не смешивать это с моей колодой
3. WHEN строится live payload THEN карты моей колоды и карты, разыгранные оппонентом, SHALL находиться в двух разных массивах и двух разных view-model
4. WHEN отображается правый список THEN он SHALL содержать только opponent played cards, а не hidden opponent hand/deck
5. WHEN матч завершается и начинается новый THEN opponent played cards SHALL очищаться для новой игры

### Requirement 5: Card row presentation and tooltip

**User Story:** Как пользователь overlay, я хочу, чтобы каждая карта была компактной строкой с art background и tooltip по наведению, чтобы быстро читать информацию, не открывая внешний справочник.

#### Acceptance Criteria

1. WHEN рендерится строка карты THEN ее размер SHALL по умолчанию использовать CSS variables с базовыми значениями около `300px` по ширине и `20px` по высоте
2. WHEN требуется изменить размеры THEN они SHALL настраиваться через design tokens, а не через hardcoded inline values по всему приложению
3. WHEN для карты доступно изображение THEN строка SHALL использовать его как background/art strip
4. WHEN на карточку наводят курсор в inspect mode THEN клиент SHALL показывать tooltip/popover с названием, описанием и изображением карты
5. IF изображение карты недоступно THEN tooltip SHALL показывать как минимум название и описание карты
6. WHEN описание карты содержит markup/служебный текст THEN UI SHALL использовать нормализованный текст, пригодный для чтения

### Requirement 6: Local live transport between CLI and client

**User Story:** Как разработчик, я хочу иметь четкий transport contract между `hs-analyzer-cli` и `hs-analyzer-client`, чтобы overlay не зависел от неявного чтения файлов или внутренних структур Go-кода.

#### Acceptance Criteria

1. WHEN `hs-analyzer-client` получает live данные THEN они SHALL приходить через versioned local contract
2. WHEN клиент подключается к live runtime THEN он SHALL уметь получить initial snapshot и затем incremental updates
3. WHEN transport реализуется THEN preferred realtime channel SHALL поддерживать push-модель, а не только polling файловой системы
4. WHEN CLI модифицируется ради UI THEN эти изменения SHALL быть допустимой частью roadmap
5. WHEN backend недоступен THEN local overlay transport SHALL продолжать работать независимо от server-side history/analytics
6. WHEN transport payload описывается THEN он SHALL содержать отдельно `player_deck_remaining[]` и `opponent_played_cards[]`

### Requirement 7: Dashboard / main application shell

**User Story:** Как пользователь desktop-приложения, я хочу при открытии клиента видеть список моих колод и детальную статистику по выбранной колоде, чтобы использовать приложение не только как overlay, но и как dashboard.

#### Acceptance Criteria

1. WHEN пользователь открывает приложение THEN он SHALL видеть левую колонку со списком всех доступных колод
2. WHEN пользователь выбирает колоду THEN main content SHALL показать детальную статистику по ней
3. WHEN открывается deck detail THEN UI SHALL показывать как минимум победы, поражения, winrate, matchup breakdown и список карт колоды
4. WHEN backend или локальный cache не готовы THEN UI SHALL иметь явные `loading`, `empty` и `error` состояния
5. WHEN deck list и deck detail отображаются THEN визуальная композиция SHALL соответствовать desktop dashboard паттерну, аналогичному приложенному референсу

### Requirement 8: Matchup analytics visualization

**User Story:** Как пользователь, я хочу видеть, против каких классов или архетипов моя колода играет лучше и хуже, чтобы быстрее оценивать эффективность деки.

#### Acceptance Criteria

1. WHEN для выбранной колоды доступны аналитические данные THEN UI SHALL показывать best/worst matchups
2. WHEN показывается summary THEN UI SHALL показывать агрегаты `wins`, `losses`, `total matches`, `winrate`
3. WHEN рендерится визуализация THEN она SHALL использовать internal UI primitives и/или native SVG/CSS, а не стороннюю chart UI library
4. WHEN backend в будущем расширяет аналитику до архетипов THEN contract SHALL допускать переход от class-level к archetype-level без слома экрана

### Requirement 9: Design system and styling rules

**User Story:** Как разработчик UI, я хочу единый design system на переменных и shadcn/ui, чтобы интерфейс не расползался по стилям и был удобно поддерживаемым.

#### Acceptance Criteria

1. WHEN реализуется UI THEN компонентная библиотека SHALL ограничиваться `shadcn/ui`
2. WHEN используются цвета, размеры, радиусы, тени, отступы, motion timings и overlay dimensions THEN они SHALL быть вынесены в CSS variables/design tokens
3. WHEN Tailwind используется THEN он SHALL ссылаться на design tokens, а не дублировать hardcoded значения по компонентам
4. WHEN создаются карточки, панели и tooltip-и THEN они SHALL следовать одному визуальному языку
5. WHEN дизайн реализуется THEN он SHALL быть ближе к deliberate game-companion UI, а не к дефолтному white/purple dashboard шаблону

### Requirement 10: Frontend architecture with FSD

**User Story:** Как разработчик, я хочу использовать FSD-архитектуру, чтобы renderer-код оставался масштабируемым и разделенным по ответственности.

#### Acceptance Criteria

1. WHEN создается renderer THEN он SHALL использовать `app / pages / widgets / features / entities / shared` layers
2. WHEN бизнес-сущности определяются THEN deck, match, live-overlay-card, tooltip-card и matchup summary SHALL иметь явные entity models
3. WHEN Electron main/preload код создается THEN он SHALL быть отделен от renderer FSD-слоев
4. WHEN shared layer создается THEN он SHALL содержать только truly shared primitives, tokens, API clients и utils
5. WHEN feature или widget зависит от transport/backend THEN границы зависимостей SHALL быть явно определены и тестируемы

### Requirement 11: Quality, testing and operational reliability

**User Story:** Как владелец проекта, я хочу принимать клиентскую подсистему только по тестам и наблюдаемому поведению, чтобы overlay и dashboard не считались завершенными без подтвержденной стабильности.

#### Acceptance Criteria

1. WHEN реализуется любой пункт roadmap THEN для него SHALL быть определены и реализованы соответствующие тесты
2. WHEN overlay получает live updates THEN UI SHALL обновляться без заметной деградации отзывчивости
3. WHEN local live transport временно пропадает THEN клиент SHALL уметь показать offline/degraded state и переподключиться
4. WHEN меняются design tokens, preload bridge, overlay contract или dashboard data contract THEN regression tests SHALL подтверждать отсутствие непреднамеренных поломок
5. WHEN этап принимается THEN обязательным критерием приемки SHALL быть успешное прохождение заранее определенного набора unit, integration, contract, visual regression и end-to-end tests
