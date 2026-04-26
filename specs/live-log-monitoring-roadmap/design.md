# Design Document: Live Log Monitoring

## Overview

Текущий `hsanalyzer` работает как статическая one-shot программа:

- читает набор логов
- парсит их полностью
- сортирует события
- сегментирует их на матчи
- формирует итоговое состояние и output

Для рабочей эксплуатации этого недостаточно. Нужен long-running режим, который:

- находит `Power.log` и связанные логи Hearthstone
- постоянно наблюдает за их изменением
- определяет активную игровую сессию
- после каждой завершенной игры сохраняет матч в каноническое хранилище

## Scope

Подсистема `live-log-monitoring-roadmap` покрывает:

- поиск и выбор актуальных логов Hearthstone
- непрерывное наблюдение за `Power.log`
- выделение active session / completed match
- финализацию матча после его завершения
- передачу завершенного матча в каноническое match storage
- CLI режим постоянного наблюдения
- observability и recovery для long-running процесса

Подсистема не покрывает на этапе архитектурной фиксации:

- окончательный выбор конкретной стратегии watcher-а
- окончательную реализацию storage layer
- финальную graph/analytics модель

## Architectural Decisions

### 1. Live-monitoring не заменяет one-shot analyze

Система должна поддерживать два режима:

- batch/one-shot analyze
- long-running live-monitoring

Это означает:

- существующий `Analyzer` остается валидным batch-компонентом
- live-monitoring строится поверх текущих parser/analyzer концепций, а не вместо них
- обе модели должны давать согласованный результат на одном и том же наборе логов

### 2. Primary stream для live режима

Основным live-источником считается `Power.log`.

Соседние логи:

- `Decks.log`
- `LoadingScreen.log`
- `LoadingScreen_old.log`
- `Zone.log`

рассматриваются как дополнительные источники контекста, которые должны быть обнаружены и использованы, если доступны.

Важно:

- `Power.log` является логом текущей активной клиентской сессии
- `Power_old.log` относится к предыдущей клиентской сессии и не должен использоваться как primary live stream
- аналогично, другие файлы вида `*_old.log` считаются архивными файлами предыдущих клиентских сессий
- наличие `*_old.log` рядом с активными файлами не должно приводить к смешению старой и текущей сессии

### 3. Разделение live-подсистемы на слои

Архитектура live-monitoring режима делится на следующие роли:

#### Path Discovery Layer

Отвечает за:

- поиск `Power.log`
- поиск соседних логов
- разрешение конфигурируемого и auto-discovered пути
- различение активных логов и архивных `*_old.log`

Приоритет источников пути должен быть таким:

1. явный путь, переданный через CLI
2. явный путь, заданный в конфигурации
3. auto-discovery стандартных директорий Hearthstone

Ключевое правило:

**если разработчик явно указал путь при запуске, система не должна молча игнорировать его в пользу auto-discovery.**

#### Watch / Live Reader Layer

Отвечает за:

- отслеживание изменений файла
- чтение только новых строк
- устойчивость к EOF, truncate, recreate и reopen

#### Session Detection Layer

Отвечает за:

- определение начала активного матча
- накопление событий активной сессии
- определение завершения матча

#### Finalization / Persistence Layer

Отвечает за:

- сбор финального представления матча
- передачу матча в каноническое match storage
- retry/idempotency защиту при сохранении

#### CLI / Observability Layer

Отвечает за:

- запуск long-running режима
- статусный вывод
- structured logging
- graceful shutdown

## Integration With Current Analyzer

Текущая реализация analyzer уже умеет:

- читать логи через `logreader`
- парсить их через registry парсеров
- сортировать события
- сегментировать общий поток в `splitEventsIntoMatches`
- строить `state.State`

Но текущий подход рассчитан на чтение всего файла целиком. Это видно по:

- [internal/analyzer/analyzer.go](/D:/self/hs-analyzer/internal/analyzer/analyzer.go)
- [internal/logreader/reader.go](/D:/self/hs-analyzer/internal/logreader/reader.go)
- [internal/logreader/file_reader.go](/D:/self/hs-analyzer/internal/logreader/file_reader.go)

Следовательно, live-monitoring не должен ломать текущий `FileReader`, а должен добавить новый live-oriented слой над ним или рядом с ним.

Ключевое решение:

**batch reader и live reader должны быть разными контрактами или разными реализациями, даже если у них общие части.**

## Active Session Model

На архитектурном уровне live-мониторинг должен различать как минимум состояния:

- `idle`
- `searching_or_waiting`
- `active_match`
- `finalizing_match`
- `persisting_match`
- `degraded`

Смысл:

- `idle` — мониторинг запущен, но матч еще не активен
- `searching_or_waiting` — лог найден, система ожидает новые строки/новый матч
- `active_match` — идет накопление событий активной игры
- `finalizing_match` — матч завершен, но еще строится финальный aggregate/state
- `persisting_match` — матч записывается в storage
- `degraded` — watcher/storage/path находятся в проблемном состоянии, но процесс еще жив

При этом active session всегда должна строиться из активного `Power.log`, а не из `Power_old.log`. Архивные `*_old.log` могут использоваться только как явный объект offline/batch анализа или как вспомогательный материал для recovery-логики, если это будет отдельно разрешено последующими решениями roadmap.

## Match Finalization Policy

После завершения матча live-система должна:

1. закрыть текущую active session
2. построить match aggregate или `state.State`
3. записать матч в каноническое хранилище
4. не допустить двойной записи того же матча
5. вернуться в режим ожидания следующего матча

Ключевое правило:

**live-monitoring не должен завершаться после одного матча.**

## Runtime Requirements Baseline

До выбора конкретной стратегии мониторинга должны быть зафиксированы требования к runtime-поведению long-running процесса.

### 1. Непрерывность работы

Live-monitoring процесс должен:

- запускаться один раз и жить до явного завершения
- не завершаться при достижении текущего EOF в `Power.log`
- корректно оставаться в режиме ожидания между матчами
- выдерживать длинные периоды без новых строк без утечек состояния и деградации управления процессом

### 2. Graceful Shutdown

Live-monitoring процесс должен:

- обрабатывать штатное завершение пользователя
- освобождать file handles, watcher resources и внутренние goroutine/resources
- завершать in-flight операции сохранения по определенным правилам
- переводить незавершенную активную сессию в согласованное состояние по policy следующих этапов roadmap

На этапе `2.1` точная политика сохранения incomplete session еще не выбирается, но обязанность graceful shutdown фиксируется как базовое требование.

### 3. Recovery After Restart

После перезапуска live-monitoring процесса или клиента Hearthstone система должна допускать восстановление работы без ручной пересборки окружения.

Это означает, что последующие решения должны учитывать:

- повторное обнаружение актуального `Power.log`
- корректную реакцию на появление нового `Power.log` рядом с `Power_old.log`
- возможность восстановить наблюдение без полной повторной обработки архивных логов как активной сессии

### 4. Resource Envelope

Выбор watcher-стратегии в следующих пунктах должен подчиняться ограничению на разумное потребление ресурсов long-running процессом.

На архитектурном уровне фиксируются такие требования:

- режим не должен требовать постоянного full reread логов
- режим не должен строиться на агрессивном busy-wait
- частота polling/watch должна быть достаточно низкой для штатной фоновой работы, но достаточной для практической near-real-time реакции
- решение должно учитывать Windows как основную среду эксплуатации

Точные численные пределы CPU, памяти и latency будут уточняться в performance acceptance criteria следующих этапов, но требование к bounded resource usage фиксируется уже сейчас.

### 5. Runtime Diagnostics

Long-running процесс должен иметь наблюдаемые runtime состояния хотя бы для:

- watcher / live reader
- active session lifecycle
- persistence pipeline
- degraded/recovery mode

Это требование нужно не только для пользователя, но и для сравнения стратегий в `2.2`: стратегия, которую трудно диагностировать в эксплуатации, не считается хорошей даже при формальной работоспособности.

### 6. Benchmark and Soak-Test Baseline

Для выбора runtime-стратегии должны существовать критерии проверки:

- short benchmark на реакцию на append новых строк
- benchmark на расходы CPU/памяти в idle режиме
- soak test на длительное ожидание без новых матчей
- soak test на серию последовательных матчей без рестарта процесса
- regression scenario на truncate/recreate/reopen

На этапе `2.1` эти критерии фиксируются как обязательный baseline, а не как уже реализованные тесты.

## Launch Model Decision

Для первого production-ready live-monitoring режима выбирается следующая модель запуска:

- отдельный long-running foreground CLI process
- один process lifecycle
- внутренние coordinator / worker loops под общим `context.Context`
- без обязательной Windows service/daemon инсталляции на первом этапе

### Почему выбрана именно эта модель

Она лучше всего соответствует текущему состоянию проекта:

- проект уже является CLI-приложением
- проще интегрировать с текущим `cobra`-слоем
- проще дебажить локально
- проще реализовать structured logging и status output
- проще контролировать graceful shutdown через единый lifecycle

### Какие варианты рассмотрены и почему не выбраны как primary

#### 1. Foreground CLI process

Плюсы:

- минимальная операционная сложность
- естественная интеграция с текущим CLI
- простой debug loop
- простой controlled shutdown

Минусы:

- зависит от того, что пользователь или оболочка держит процесс запущенным

Решение:

- выбрать как primary launch model

#### 2. Background daemon / Windows service

Плюсы:

- удобен для always-on production эксплуатации

Минусы:

- выше стоимость реализации и отладки
- усложняет lifecycle, инсталляцию и recovery уже на раннем этапе
- преждевременно для текущей стадии проекта

Решение:

- не выбирать как первую обязательную модель
- оставить как возможное расширение после стабилизации live-monitoring ядра

## Monitoring Strategy Decision

Для live-monitoring `Power.log` выбирается **hybrid strategy**:

- low-frequency directory/file polling как reliable baseline
- событийный wake-up слой (`fsnotify`-style watch) как ускоритель реакции, если доступен и стабилен
- offset-based incremental reading как обязательный механизм чтения
- reopen/restat logic как обязательный recovery-механизм

Иначе говоря:

- источником истины для того, что именно читать дальше, является не сам watch event
- источником истины являются текущие file identity/size/offset checks и логика incremental reader-а

### Почему не выбран pure file watch

Чистый `fsnotify`-подход не выбран как primary strategy, потому что:

- file watch события могут дублироваться или теряться
- rename/recreate поведения на Windows сложнее и менее предсказуемы
- `Power.log -> Power_old.log` + создание нового `Power.log` требует явной restat/reopen логики в любом случае
- watcher-событие само по себе не гарантирует, что reader корректно понял текущую file identity

Вывод:

- watch event полезен как триггер ускоренного пересмотра состояния
- но не должен быть единственным механизмом истины

### Почему не выбран pure polling

Чистый polling тоже не выбран как единственная стратегия, потому что:

- при слишком частом polling ненужно растет idle CPU cost
- при слишком редком polling ухудшается latency реакции на новые строки
- отсутствует быстрый wake-up при активной записи, если полагаться только на fixed interval

Вывод:

- polling нужен как reliability baseline
- но его лучше усиливать event-driven wake-up сигналом

### Почему не выбран tail-like как отдельное решение

Tail-like модель полезна как reader abstraction, но сама по себе не отвечает на вопросы:

- как различать старый и новый `Power.log`
- как переживать rename/recreate
- как реагировать на `_old` ротацию
- как заново открыть файл при смене file identity

Поэтому tail-like behavior должен быть частью incremental reader-а, а не полной стратегией мониторинга.

### Почему hybrid является лучшим компромиссом

Hybrid strategy дает:

- надежность polling baseline
- приемлемую near-real-time реакцию за счет watch wake-up
- контролируемое поведение на Windows
- явную поддержку truncate/recreate/reopen
- меньшую зависимость от особенностей конкретного file watch backend

## Monitoring Strategy Matrix

### 1. Pure file watch

- Надежность при rename/recreate: средняя или нестабильная
- Idle cost: низкая
- Реакция на append: быстрая
- Простота recovery: средняя
- Пригодность для Windows-first: ограниченная

### 2. Pure polling

- Надежность при rename/recreate: высокая
- Idle cost: зависит от интервала
- Реакция на append: зависит от интервала
- Простота recovery: высокая
- Пригодность для Windows-first: высокая

### 3. Tail-like only

- Надежность как complete monitoring strategy: недостаточная
- Reader abstraction usefulness: высокая
- Recovery coverage: недостаточная без дополнительной логики

### 4. Hybrid watch + polling + offset reader

- Надежность при rename/recreate: высокая
- Idle cost: контролируемая
- Реакция на append: быстрая
- Простота recovery: высокая
- Пригодность для Windows-first: высокая

Решение:

- выбрать `hybrid watch + polling + offset reader` как рекомендуемую стратегию

## Runtime Validation Criteria For Strategy Selection

Стратегия считается приемлемой только если последующая реализация сможет пройти такие сценарии:

- append новых строк без reread всего файла
- длительный idle период без роста CPU и memory drift
- `Power.log` rename в `Power_old.log` и появление нового `Power.log`
- truncate существующего файла
- recreate файла после временного исчезновения
- несколько матчей подряд в одном живом процессе
- graceful shutdown во время idle и во время in-flight обработки

Эти критерии должны лечь в unit/integration/regression/soak tests следующих этапов roadmap.

## Dependency on Match Storage Roadmap

Эта подсистема архитектурно зависит от `match-storage-roadmap`.

После завершения матча live-поток должен сохранять его:

- не только в materialized output
- а именно в каноническое match storage

То есть `timeline_*` не должен быть обязательным финальным результатом live режима.

## CLI Policy

Live-monitoring должен быть отдельным, явно различимым режимом CLI.

Допустимые варианты дальнейшей реализации:

- новый subcommand
- `analyze --watch`
- отдельный daemon-like command

На этапе пункта 1 фиксируется только требование:

**режим постоянного мониторинга должен быть отдельной, явной пользовательской возможностью CLI.**

Дополнительно для CLI фиксируется правило:

- разработчик должен иметь возможность явно передать путь к логам или активному `Power.log`
- этот путь должен иметь приоритет над auto-discovery
- ошибочный явный путь должен приводить к явной ошибке, а не к тихому fallback

## Recovery and Fault Tolerance

Live-monitoring должен проектироваться с учетом типовых реальных проблем:

- `Power.log` временно недоступен
- Hearthstone перезапущен
- лог truncated или recreated
- текущий `Power.log` переведен в `Power_old.log`, а новая клиентская сессия создала новый `Power.log`
- storage временно недоступен
- сам monitor процесс был остановлен и запущен снова

Это не означает, что вся recovery логика уже реализована, но архитектурно она должна быть предусмотрена и выделена в roadmap как обязательная часть, а не как post-factum патч.

## Testing and Acceptance

Пункт 1 roadmap считается принятым, когда:

1. определен scope live-monitoring подсистемы
2. зафиксировано сосуществование batch и live режимов
3. описаны роли path discovery, watch/live reader, session detection, finalization/persistence и CLI/observability layers
4. зафиксировано, что завершенный матч должен попадать в каноническое match storage
5. определены acceptance criteria, по которым последующие реализации можно проверять тестами

На следующих этапах каждое решение должно подтверждаться:

- unit tests
- integration tests
- regression tests
- soak tests
- performance tests

Без успешно пройденных тестов задача не считается принятой.

Пункт `2.1` roadmap считается принятым, когда:

1. зафиксированы требования к непрерывной работе long-running процесса
2. зафиксированы требования к graceful shutdown
3. зафиксированы требования к recovery после перезапуска
4. зафиксированы ограничения на bounded resource usage
5. определен baseline benchmark и soak-test criteria для следующего шага выбора watcher-стратегии

Пункт `2.2` roadmap считается принятым, когда:

1. рассмотрены pure watch, pure polling, hybrid и tail-like подходы
2. выбран preferred launch model для long-running процесса
3. выбрана рекомендуемая стратегия мониторинга файлов
4. зафиксированы причины отказа от альтернатив как primary strategy
5. определен test baseline для проверки выбранной стратегии
