# Requirements Document: Live Log Monitoring

## Introduction

Подсистема `live-log-monitoring-roadmap` определяет переход `hsanalyzer` от одноразового анализа логов к long-running режиму, который после запуска непрерывно наблюдает за логами Hearthstone, определяет активную игровую сессию и сохраняет каждый завершенный матч в каноническое хранилище.

Текущая версия проекта в основном работает как batch/one-shot analyzer: она один раз читает набор файлов, строит массив событий и материализует результат. Новая подсистема должна добавить режим постоянного наблюдения без разрушения уже существующего `analyze` flow.

Корректно выполненные тесты являются обязательной основой приемки для всех требований этой подсистемы.

## Requirements

### Requirement 1: Непрерывный режим работы

**User Story:** Как пользователь, я хочу запустить программу один раз и оставить ее работать постоянно, чтобы она сама отслеживала новые матчи Hearthstone.

#### Acceptance Criteria

1. WHEN программа запускается в live-monitoring режиме THEN она SHALL продолжать работу до явного завершения пользователем или фатальной ошибки
2. WHEN live-monitoring режим активен THEN система SHALL ожидать новые записи в логах без повторного запуска процесса
3. WHEN программа завершается штатно THEN она SHALL освобождать ресурсы watcher-ов и reader-ов корректно
4. WHEN live-monitoring работает длительное время THEN система SHALL не требовать периодического ручного перезапуска для штатной эксплуатации
5. WHEN система не получает новых строк в течение длительного времени THEN она SHALL оставаться в корректном режиме ожидания, а не завершаться как batch-процесс

### Requirement 1.1: Runtime ограничения long-running режима

**User Story:** Как разработчик и оператор, я хочу заранее определить runtime-требования для long-running процесса, чтобы дальнейший выбор watcher-стратегии и recovery-механики был измеримым и проверяемым.

#### Acceptance Criteria

1. WHEN проектируется live-monitoring режим THEN система SHALL иметь зафиксированные требования к непрерывной работе после запуска
2. WHEN проектируется live-monitoring режим THEN система SHALL иметь зафиксированные требования к graceful shutdown
3. WHEN проектируется live-monitoring режим THEN система SHALL иметь зафиксированные требования к восстановлению после перезапуска приложения или Hearthstone
4. WHEN сравниваются runtime стратегии THEN система SHALL иметь зафиксированные ограничения по CPU, памяти и latency обработки новых строк
5. WHEN оценивается поведение live-monitoring режима THEN система SHALL иметь определенный набор benchmark и soak-test criteria

### Requirement 1.2: Модель запуска long-running процесса

**User Story:** Как разработчик, я хочу иметь явную модель запуска live-monitoring режима, чтобы процессом было просто управлять, диагностировать и завершать его без отдельной сервисной инфраструктуры.

#### Acceptance Criteria

1. WHEN выбирается модель запуска live-monitoring режима THEN система SHALL иметь явно зафиксированный preferred launch model
2. WHEN live-monitoring запускается пользователем из CLI THEN preferred launch model SHALL работать как долгоживущий foreground process
3. WHEN preferred launch model использует внутренние worker/supervisor loops THEN они SHALL быть управляемы через общий lifecycle и graceful shutdown
4. WHEN сравниваются альтернативы launch model THEN выбранная модель SHALL быть обоснована простотой эксплуатации, дебага и интеграции с текущим CLI

### Requirement 3.1: Выбор стратегии мониторинга файлов

**User Story:** Как разработчик, я хочу выбрать основную стратегию отслеживания изменений `Power.log`, чтобы live-monitoring был устойчивым на Windows и не зависел от хрупкого единственного механизма наблюдения.

#### Acceptance Criteria

1. WHEN сравниваются file monitoring стратегии THEN система SHALL оценивать `file system watch`, `polling`, `hybrid` и `tail-like` подходы
2. WHEN выбирается стратегия THEN она SHALL учитывать `truncate`, `recreate`, `rename to *_old`, idle waiting и delayed writes
3. WHEN выбирается стратегия THEN она SHALL учитывать Windows как основную среду эксплуатации
4. WHEN стратегия утверждена THEN она SHALL иметь зафиксированные acceptance tests для append, idle, truncate, recreate, rename и reopen сценариев

### Requirement 2: Обнаружение логов Hearthstone

**User Story:** Как пользователь, я хочу, чтобы система находила `Power.log` и связанные логи Hearthstone автоматически или по явному пути, чтобы не настраивать всё вручную в каждом запуске.

#### Acceptance Criteria

1. WHEN пользователь задает путь к логам явно THEN система SHALL использовать его как приоритетный источник
2. WHEN явный путь не задан THEN система SHALL уметь искать `Power.log` в стандартных или auto-discovered путях
3. WHEN `Power.log` найден THEN система SHALL уметь определить соседние `Decks.log`, `LoadingScreen.log`, `LoadingScreen_old.log` и `Zone.log`
4. IF путь или лог не найден THEN система SHALL вывести диагностически полезное сообщение об ошибке
5. WHEN в каталоге одновременно присутствуют `Power.log` и `Power_old.log` THEN система SHALL считать `Power.log` активным логом текущей сессии, а `Power_old.log` - архивом предыдущей
6. WHEN в каталоге присутствуют другие `*_old.log` файлы THEN система SHALL не использовать их как primary live stream активной сессии

### Requirement 2.1: Явный путь как developer override

**User Story:** Как разработчик, я хочу иметь возможность при запуске программы задать точный путь к каталогу логов или к конкретному активному логу, чтобы не зависеть только от auto-discovery.

#### Acceptance Criteria

1. WHEN разработчик передает явный путь при запуске программы THEN система SHALL использовать его как приоритетный override над auto-discovery
2. WHEN передан путь к каталогу логов THEN система SHALL искать активный `Power.log` и связанные логи внутри этого каталога
3. WHEN передан путь к конкретному `Power.log` THEN система SHALL использовать именно этот файл как активный источник мониторинга
4. IF переданный явный путь некорректен THEN система SHALL завершаться с понятной диагностической ошибкой, а не молча переключаться на auto-discovery
5. WHEN используется явный путь THEN система SHALL явно отражать это в статусном выводе и structured logs

### Requirement 3: Наблюдение за растущим `Power.log`

**User Story:** Как разработчик, я хочу читать только новые строки из растущего `Power.log`, чтобы не перечитывать весь файл и не терять события.

#### Acceptance Criteria

1. WHEN `Power.log` увеличивается THEN система SHALL читать только новые строки
2. WHEN достигнут текущий EOF THEN система SHALL ждать новых данных вместо завершения процесса
3. WHEN файл был truncate/recreate/reopen THEN система SHALL корректно продолжить наблюдение по правилам live reader-а
4. WHEN возникают временные I/O проблемы THEN система SHALL пытаться восстановиться без немедленного падения процесса, если это безопасно

### Requirement 4: Определение активной игровой сессии

**User Story:** Как пользователь, я хочу, чтобы система понимала, идет ли сейчас матч, чтобы она сохраняла только законченные катки и не путала игровые сессии.

#### Acceptance Criteria

1. WHEN новые события начинают соответствовать старту матча THEN система SHALL открыть активную игровую сессию
2. WHEN матч продолжается THEN система SHALL накапливать относящиеся к нему события
3. WHEN матч завершен THEN система SHALL закрыть активную сессию и передать матч на финализацию
4. WHEN подряд сыграно несколько матчей THEN система SHALL корректно разделять их на отдельные сессии
5. WHEN в каталоге логов остаются файлы предыдущих сессий с суффиксом `_old` THEN система SHALL не смешивать их содержимое с активной сессией из текущего `Power.log`

### Requirement 5: Сохранение каждого завершенного матча

**User Story:** Как владелец данных, я хочу, чтобы каждая завершенная катка сохранялась в БД автоматически, чтобы дальше использовать ее для аналитики и реконструкции.

#### Acceptance Criteria

1. WHEN матч завершен THEN система SHALL сохранять его в каноническое match storage
2. WHEN матч уже был сохранен THEN система SHALL избегать дублирующей записи
3. WHEN storage временно недоступен THEN система SHALL выполнять согласованную retry/idempotency стратегию
4. WHEN сохранение прошло успешно THEN система SHALL иметь возможность в дальнейшем реконструировать матч из storage

### Requirement 6: Совместимость с текущим analyzer

**User Story:** Как разработчик, я хочу переиспользовать текущие `Analyzer`, `Parser`, `Pipeline` и модели событий, чтобы live-режим не становился отдельной полностью несвязанной системой.

#### Acceptance Criteria

1. WHEN реализуется live-monitoring режим THEN он SHALL переиспользовать существующие parser/analyzer компоненты настолько, насколько это разумно
2. WHEN матч финализируется THEN система SHALL иметь определенный путь преобразования накопленных данных в `state.State` или эквивалентный match aggregate
3. WHEN batch и live режимы сосуществуют THEN их поведение SHALL оставаться согласованным на одинаковом наборе логов

### Requirement 7: CLI режим постоянного наблюдения

**User Story:** Как пользователь CLI, я хочу явно запускать режим постоянного мониторинга, чтобы отличать его от одноразового анализа.

#### Acceptance Criteria

1. WHEN пользователь запускает live-monitoring режим THEN CLI SHALL активировать long-running наблюдение за логами
2. WHEN режим запущен THEN CLI SHALL выводить полезный статус: найденные логи, активная сессия, завершение матча, результат записи
3. WHEN пользователь завершает процесс THEN система SHALL выполнять graceful shutdown

### Requirement 8: Наблюдаемость и диагностика

**User Story:** Как пользователь и разработчик, я хочу видеть, что live-monitoring работает корректно, чтобы быстро понимать, где именно возникла проблема: в поиске логов, в watcher-е, в session detector или в storage.

#### Acceptance Criteria

1. WHEN live-monitoring режим работает THEN система SHALL писать структурированные диагностические логи о watcher-е, session detector и persistence pipeline
2. WHEN система деградирует THEN она SHALL уметь сообщить, что именно деградировало: path discovery, log reader, active session или storage
3. WHEN пользователь наблюдает за процессом THEN он SHALL иметь минимальный понятный status output
4. WHEN система находится в long-running режиме THEN она SHALL иметь наблюдаемые runtime состояния для watcher-а, active session и persistence pipeline

### Requirement 9: Отказоустойчивость и восстановление

**User Story:** Как пользователь, я хочу, чтобы мониторинг переживал типовые сбои, например временное отсутствие логов или перезапуск Hearthstone, чтобы не терять целые серии матчей.

#### Acceptance Criteria

1. WHEN `Power.log` временно исчезает или заменяется THEN система SHALL переходить в восстановимый режим и пытаться продолжить работу
2. WHEN Hearthstone перезапущен THEN система SHALL уметь повторно обнаружить актуальные логи и вернуться к monitoring flow
3. WHEN приложение мониторинга было перезапущено THEN архитектура SHALL допускать восстановление monitoring state по принятым правилам
4. WHEN текущий `Power.log` переносится в `Power_old.log` при старте новой клиентской сессии THEN система SHALL корректно переопределять активный источник мониторинга без повторной обработки архивного лога как новой активной сессии

### Requirement 10: Тестируемость и приемка

**User Story:** Как владелец проекта, я хочу принимать live-monitoring функциональность только по подтвержденным тестам, чтобы long-running логика не считалась завершенной без проверки устойчивости.

#### Acceptance Criteria

1. WHEN реализуется любой пункт roadmap THEN для него SHALL быть определены соответствующие тесты
2. WHEN реализация считается завершенной THEN ее тесты SHALL быть реализованы и успешно проходить
3. WHEN меняется watcher, session detector, finalization logic или persistence flow THEN regression tests SHALL подтверждать отсутствие непреднамеренных изменений
4. WHEN принимается этап live-monitoring подсистемы THEN критерием приемки SHALL быть успешное прохождение заранее определенного набора unit, integration, regression, soak и performance tests

### Requirement 11: Live card stats projection для клиента

**User Story:** Как пользователь внешнего клиента, я хочу в live runtime видеть, какие карты уже сыграл я и какие сыграл оппонент, чтобы UI мог показывать компактную и полезную статистику матча без полного replay.

#### Acceptance Criteria

1. WHEN live-monitoring получает `CardPlayedEvent` THEN система SHALL обновлять lightweight client projection без ожидания завершения матча
2. WHEN карта сыграна игроком или оппонентом THEN система SHALL агрегировать статистику отдельно по bucket-ам `self` и `opponent`
3. WHEN запись о карте публикуется клиенту THEN она SHALL содержать как минимум `name`, `description` и `image`
4. WHEN текст карты берется из card database THEN система SHALL публиковать очищенный человекочитаемый `description`, а не сырой HTML-marked text
5. WHEN изображение карты недоступно через локальную card database THEN система SHALL использовать отдельную image resolution policy с явным fallback
6. WHEN активный матч начинается заново THEN live card stats projection SHALL быть сброшена для нового match lifecycle
7. WHEN матч завершается THEN система SHALL сохранять финальный snapshot до старта следующего матча или явного reset
8. WHEN monitor process перезапускается в середине матча THEN система SHALL либо честно восстановить live projection, либо явно пометить ее как partial/incomplete, но не SHALL silently corrupt statistics
