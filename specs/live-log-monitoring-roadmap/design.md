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
