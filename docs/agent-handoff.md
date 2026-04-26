# Agent Handoff

Этот документ нужен для быстрого включения второго агента в проект `hsanalyzer` без чтения всей истории чата.

## 1. Что это за проект

`hsanalyzer` — CLI-инструмент на Go для разбора логов Hearthstone, восстановления матчей и экспорта результатов в:

- `text`
- `json` (`human` и `replay`)

Текущая кодовая база уже умеет:

- читать логи Hearthstone
- парсить события из `Power.log`, `Decks.log`, `LoadingScreen.log` / `LoadingScreen_old.log`, `Zone.log`
- объединять события через `Analyzer` / `Pipeline`
- сегментировать общий набор логов на массив матчей
- строить текстовый и JSON-вывод

Главная текущая проблема:

- `output/timeline_*` слишком тяжелый
- один файл хранит сразу несколько матчей
- JSON многократно дублирует `card.info`, snapshot-like структуры и повторяющиеся данные
- это плохо подходит для будущего анализа решений и “лучшего хода”

## 2. Что уже решено архитектурно

В проекте создана новая спецификация:

- [specs/match-storage-roadmap/tasks.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/tasks.md)
- [specs/match-storage-roadmap/requirements.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/requirements.md)
- [specs/match-storage-roadmap/design.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/design.md)

Уже зафиксированы следующие решения:

1. Единица хранения — один завершенный матч.
2. `timeline_*.json` больше не считается source of truth.
3. Каноническое match storage должно быть primary storage.
4. Human/debug/compatibility outputs должны материализоваться из канонического хранения.
5. Графовая модель рассматривается как secondary projection для аналитики, а не как обязательное первое source-of-truth хранилище.
6. Навигационная структура матча должна быть `match -> turns -> event refs`.
7. Архитектура должна быть пригодна для будущих `decision points`, `candidate actions`, `state -> action -> outcome`.

## 3. Что уже выполнено

В roadmap выполнен:

- `[x] 1. Сформировать архитектурное решение для нового слоя хранения матчей`

Это уже отражено в `design.md` и `requirements.md`.

Последние важные коммиты:

- `adb05a4` — roadmap для хранения матчей и аналитики
- `ee2bab2` — архитектурная спецификация хранения матчей

Оба коммита уже запушены в `origin/main`.

## 4. Что сейчас НЕ сделано

Следующий активный блок — пункт `2` из roadmap:

- исследовать и выбрать primary storage

Внутри него ближайший логичный старт:

- `2.1 Сформировать критерии выбора хранилища`

Нужно не сразу писать runtime-код, а сначала зафиксировать:

- требования по размеру
- требования по скорости записи/чтения
- требования к локальному запуску
- требования к миграциям
- требования к debugging
- требования к extraction/use-cases для будущего анализа лучшего хода

После этого:

- сравнивать `SQLite`, альтернативы и strategy для graph projection

## 5. Что важно понимать про будущую архитектуру

### Каноническое хранение

Ожидаемая модель:

- `match`
- `turn`
- `event`
- позже: `entity`, `snapshot`, `decision_point`, `candidate_action`

### Дерево

Нельзя хранить только “какие карты были сыграны на ходе”.

Нужно хранить именно:

- `match -> turns -> event refs`

Потому что для будущего анализа решений важны не только розыгрыши карт, но и:

- draw
- attack
- damage
- heal
- death
- discover
- effect_triggered
- zone/resource transitions

### Replay / восстановление

Новая модель должна быть event-sourced:

- каноническое событие
- редкие snapshots/checkpoints
- реконструкция состояния по запросу

### Анализ лучшего хода

Roadmap проекта идет в сторону:

- анализа ситуаций
- поиска “лучшего хода”
- explainability
- возможной графовой проекции для `state -> action -> outcome`

Поэтому нельзя проектировать storage только как архив логов. Он должен быть пригоден для decision analysis.

## 6. Почему не выбран “чистый graph-first”

В обсуждении был важный вывод:

- `SQLite + graph DB` вместе — допустимая и сильная архитектура
- но граф не должен автоматически становиться primary storage

Рационал:

- primary storage должен быть простым для записи и миграций
- materialized graph можно перестраивать
- graph schema почти наверняка будет меняться быстрее, чем canonical match schema

На текущем этапе зафиксировано именно это направление:

- primary storage для истины
- graph projection для аналитики

Конкретная СУБД еще не выбрана. Это следующий этап roadmap.

## 7. Правила приемки

Это критично.

Для нового roadmap уже зафиксировано:

- каждый пункт должен сопровождаться тестами
- успешно выполненные тесты являются основой приемки
- задача не считается завершенной без тестов

Это правило записано в:

- [specs/match-storage-roadmap/tasks.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/tasks.md)
- [specs/match-storage-roadmap/requirements.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/requirements.md)
- [specs/match-storage-roadmap/design.md](/D:/self/hs-analyzer/specs/match-storage-roadmap/design.md)

Если другой агент продолжает работу, он должен учитывать это как gate на приемку.

## 8. Как устроен текущий проект

Ключевые точки входа:

- [README.md](/D:/self/hs-analyzer/README.md)
- [cmd/hsanalyzer](/D:/self/hs-analyzer/cmd/hsanalyzer)
- [internal/analyzer](/D:/self/hs-analyzer/internal/analyzer)
- [internal/parser](/D:/self/hs-analyzer/internal/parser)
- [internal/formatter](/D:/self/hs-analyzer/internal/formatter)
- [internal/state](/D:/self/hs-analyzer/internal/state)

Полезные текущие файлы по проблеме timeline:

- [internal/formatter/json_formatter.go](/D:/self/hs-analyzer/internal/formatter/json_formatter.go)
- [internal/formatter/replay_json_formatter.go](/D:/self/hs-analyzer/internal/formatter/replay_json_formatter.go)

Именно в этих форматтерах сейчас видно основное раздувание JSON:

- повторяющийся `card.info`
- state-like snapshots
- тяжелые вложенные структуры

## 9. Текущая ветка и статус

На момент создания этого handoff:

- рабочая ветка: `main`
- последние изменения по match storage уже в `origin/main`
- незакоммиченных изменений до создания этого файла не было

Если агент продолжает работу после этого документа, ему стоит начать с:

1. прочитать `specs/match-storage-roadmap/*`
2. свериться с `README.md`
3. взять пункт `2.1` roadmap
4. сначала оформить spec/decision layer, потом переходить к runtime-коду

## 10. Рекомендуемый следующий шаг для второго агента

Самый правильный следующий инкремент:

1. Выполнить `2.1` в `specs/match-storage-roadmap/tasks.md`
2. Зафиксировать критерии выбора primary storage
3. Подготовить сравнимую decision matrix для:
   - `SQLite`
   - файлового/append-only compact storage
   - embedded KV/других локальных вариантов
   - роли graph projection
4. Определить test/benchmark criteria для выбора

После этого можно переходить к `2.2` и уже принимать решение по primary storage.
