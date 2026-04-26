# Graph Storage Options and Integration Strategy (Task 3.2)

## Контекст

Граф не является primary source of truth. Канонические данные матчей хранятся в primary storage (см. решение по `SQLite`), а graph слой строится как вторичная проекция.

## 1. Варианты graph storage

### A. Полноценная graph DB (Neo4j/Memgraph-class)

Плюсы:

- богатый query language для графовых обходов
- нативные индексы и traversal-оптимизации
- удобна для сложных explainability-запросов

Минусы:

- дополнительная инфраструктура и devops-нагрузка
- сложнее локальный onboarding
- отдельный lifecycle миграций и бэкапов

### B. Embedded graph-подход

Примеры:

- in-process graph index
- библиотечная графовая структура поверх локальных файлов

Плюсы:

- минимальная внешняя инфраструктура
- проще запуск на developer workstation

Минусы:

- часто слабее query-возможности
- больше custom-кода для индексов и consistency
- риск роста технического долга при расширении аналитики

### C. Materialized graph поверх primary storage (SQLite-first)

Идея:

- graph projection хранится как проекционные таблицы/ребра в отдельной схеме/БД
- source данных - только canonical match storage

Плюсы:

- полностью совместимо с принятым primary storage решением
- rebuild projection детерминирован и воспроизводим
- нет зависимости от внешней graph-инфраструктуры на базовом этапе

Минусы:

- сложные graph traversal запросы могут быть медленнее, чем в dedicated graph DB
- часть graph-функциональности придется реализовать через materialization/индексы

## 2. Сравнительная оценка

Шкала `0..5`.

| Критерий | Graph DB | Embedded Graph | Materialized Graph over SQLite |
|---|---:|---:|---:|
| Локальный запуск | 2 | 4 | 5 |
| Стоимость поддержки | 2 | 3 | 4 |
| Rebuild проекции | 4 | 3 | 5 |
| Гибкость запросов | 5 | 2 | 3 |
| Интеграция с текущим проектом | 3 | 3 | 5 |
| Общий баланс для текущего этапа roadmap | 3 | 3 | 5 |

## 3. Решение по роли графа

Граф фиксируется как:

- `secondary projection` (не primary storage)
- `optional analytics layer` на ранних этапах runtime-реализации

Это означает:

- канонический путь ingest/read не зависит от графа
- аналитические graph-возможности можно включать инкрементально
- при недоступности projection слоя базовая функциональность продукта сохраняется

## 4. Рекомендуемая стратегия интеграции

Этапы:

1. Писать canonical match data в primary storage.
2. Строить graph projection из canonical IDs.
3. Поддерживать `full rebuild` проекции как штатную операцию.
4. По мере роста требований добавить optional адаптер к dedicated graph DB без смены source of truth.

## 5. Тест-критерии для rebuild и целостности

### 5.1 Rebuild consistency

- После полного удаления projection-слоя rebuild восстанавливает тот же набор узлов/ребер.

### 5.2 Referential integrity

- Каждый graph node/edge ссылается на существующие canonical IDs (`match_id`, `turn_id`, `event_id`, `decision_point_id`).

### 5.3 Incremental vs full rebuild equivalence

- Проекция, построенная инкрементально, эквивалентна проекции после full rebuild на том же наборе матчей.

### 5.4 Degradation safety

- При отключенном/неисправном projection слое запись и чтение canonical storage продолжают работать.

## 6. Критерий завершения 3.2

Шаг `3.2` считается завершенным, когда:

1. Сравнены три класса решений: dedicated graph DB, embedded graph, materialized graph.
2. Зафиксировано, что граф - secondary projection.
3. Определена стратегия интеграции с primary storage.
4. Определены тест-критерии на rebuild и целостность проекций.
