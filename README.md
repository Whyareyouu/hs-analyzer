# Hearthstone Log Analyzer

`hsanalyzer` — CLI-инструмент на Go для разбора логов Hearthstone, восстановления хронологии матчей и экспорта результата в `text` и `json`.

## Текущий статус

Сейчас в проекте реализованы:

- чтение конфигурации через `yaml` и `viper`
- загрузка базы карт из `data/cards.json`
- парсеры для `Decks.log`, `Power.log`, `LoadingScreen.log` / `LoadingScreen_old.log` и `Zone.log`
- `Analyzer` и `Pipeline` для объединения, сортировки, сегментации и применения событий
- автоматическое разбиение общего набора логов на массив матчей
- текстовый formatter с поддержкой `ru` и `en`
- два JSON-режима: `human` и `replay`
- CLI на `cobra`
- structured logging через `zap`

Ограничения текущей версии:

- `GameEndEvent` пока не восстанавливается из логов автоматически для всех сценариев
- часть эффектов `effect_triggered` уже человекочитаема, но словарь расшифровки еще не полный
- поля `filters.*` и `performance.workers` пока зарезервированы и не влияют на итоговый анализ

## Требования

- Go `1.21+`
- логи Hearthstone
- база карт `data/cards.json`

## Быстрый старт

```bash
go mod download
go build -o bin/hsanalyzer.exe ./cmd/hsanalyzer
```

Или через `Makefile`:

```bash
make install
make build
```

## Быстрый запуск desktop-приложения

Полный локальный стек можно поднять одной командой из корня репозитория:

```powershell
./scripts/run-app.ps1
```

```cmd
run-app.cmd
```

```cmd
scripts\run-app.cmd
```

```bash
bash ./scripts/run-app.sh
```

Что делают скрипты:

- при необходимости собирают `hs-analyzer-be`, `hs-analyzer-cli` и Electron client
- поднимают backend на `127.0.0.1:8080`
- запускают `hsanalyzer monitor` с live API на `127.0.0.1:8765`
- открывают `hs-analyzer-client` в foreground и завершают фоновые процессы после закрытия окна

Полезные варианты запуска:

```powershell
./scripts/run-app.ps1 -LogsPath logs
./scripts/run-app.ps1 -RebuildGo -RebuildClient
```

```bash
bash ./scripts/run-app.sh --logs-path logs
bash ./scripts/run-app.sh --rebuild-go --rebuild-client
```

По умолчанию `monitor` использует auto-discovery логов Hearthstone. Для локального прогона на репозиторных логах можно явно передать `logs`.

## Команды CLI

Общая справка:

```bash
go run ./cmd/hsanalyzer --help
```

Доступные команды:

- `analyze` — анализ логов Hearthstone
- `update-cards` — загрузка актуальной базы карт

Глобальный флаг:

- `--log-level debug|info|warn|error` — уровень structured logging в `stderr`

### analyze

Справка:

```bash
go run ./cmd/hsanalyzer analyze --help
```

Флаги:

- `--config` — путь к `yaml`-конфигу
- `--logs-path` — путь к каталогу с логами
- `--cards-db-path` — путь к `cards.json`
- `--format` — формат вывода: `text` или `json`
- `--json-mode` — режим JSON: `replay` или `human`
- `--output` — `stdout` или `file`
- `--output-file` — путь к файлу результата, если `--output file`
- `--lang` — `ru` или `en`
- `--verbose` — расширенный текстовый вывод

Примеры:

```bash
go run ./cmd/hsanalyzer analyze --logs-path logs --cards-db-path data/cards.json

go run ./cmd/hsanalyzer analyze \
  --logs-path logs \
  --cards-db-path data/cards.json \
  --format text \
  --lang ru \
  --json-mode replay

go run ./cmd/hsanalyzer analyze \
  --logs-path "C:\\Program Files (x86)\\Hearthstone\\Logs" \
  --cards-db-path data/cards.json \
  --json-mode human \
  --lang ru \
  --verbose

go run ./cmd/hsanalyzer analyze \
  --config .\configs\config.yaml \
  --format json \
  --json-mode replay \
  --output file \
  --output-file .\result.json
```

### update-cards

Справка:

```bash
go run ./cmd/hsanalyzer update-cards --help
```

Примеры:

```bash
go run ./cmd/hsanalyzer update-cards
go run ./cmd/hsanalyzer update-cards --locale enUS
go run ./cmd/hsanalyzer update-cards --locale ruRU --out data/cards.json
```

Команда загружает файл карт из HearthstoneJSON:

```text
https://api.hearthstonejson.com/v1/latest/<locale>/cards.json
```

## Конфигурация

Пример `config.yaml`:

```yaml
logs_path: logs
cards_db_path: data/cards.json

output:
  format: text
  json_mode: replay
  destination: stdout
  file_path: ""
  language: ru
  verbose: false

filters:
  game_mode: []
  min_turns: 0
  max_turns: 0

performance:
  workers: 1
  buffer_size: 65536
```

Готовый пример с комментариями лежит в [configs/config.yaml](configs/config.yaml).

Что реально используется сейчас:

- `logs_path`
- `cards_db_path`
- `output.format`
- `output.json_mode`
- `output.destination`
- `output.file_path`
- `output.language`
- `output.verbose`
- `performance.buffer_size`

Что пока не влияет на поведение analyzer:

- `filters.*`
- `performance.workers`

## Поддерживаемые логи

CLI и analyzer сейчас работают со следующими файлами:

- `Power.log`
- `Decks.log`
- `LoadingScreen.log`
- `LoadingScreen_old.log`
- `Zone.log` — опционально

Важно:

- для `LoadingScreen` используется fallback на `LoadingScreen_old.log`
- если `Zone.log` отсутствует, analyzer продолжает работу без него

## Что умеет парсить

Сейчас реализован парсинг следующих событий:

- начало игры и режим
- конец игры как доменная модель события
- список карт колоды
- начало и конец ходов
- розыгрыш карт
- добор карт
- атаки
- урон
- лечение
- смерти
- срабатывание эффектов
- изменение стоимости карт
- `Discover`

## Формат вывода

`TextFormatter` формирует:

- заголовок игры
- хронологию по ходам
- события добора и розыгрыша карт
- боевые события
- эффекты и `Discover`
- текущее состояние руки
- итоговую сводку

Поддерживаются языки:

- `ru`
- `en`

JSON-вывод поддерживает два режима:

- `human` — компактный человекочитаемый JSON для просмотра и отладки
- `replay` — подробный JSON для клиентского восстановления матча

Оба JSON-режима теперь отдают верхнеуровневый документ с массивом матчей:

```json
{
  "schema_version": "0.0.1",
  "mode": "replay",
  "match_count": 2,
  "matches": [
    {
      "game": {
        "id": "game-1",
        "mode": "STANDARD"
      },
      "turns": []
    }
  ]
}
```

`human`-режим содержит:

- краткую структуру по ходам
- `opening_hand` в начале каждого хода
- человекочитаемые события и карточную информацию
- удобный формат для чтения глазами

`replay`-режим содержит:

- `opening_state` в начале каждого хода
- `action`, `delta`, `state_before`, `state_after` для ключевых событий
- ресурсы игроков: `mana`, `max_mana`, `overload`
- зоны и их содержимое: `hand`, `board`, `graveyard`, `known_deck`, `deck_list`
- `final_state` по каждому матчу

Автосохранение timeline всегда пишет `replay`-JSON в каталог `output/` проекта с именем:

```text
timeline_<timestamp>_replay.json
```

Пример текстового вывода:

```text
Игра
ID игры: n/a
Режим: STANDARD

Ходы
 Ход 1 (Оппонент)
  🔄 Начало хода 1 (Оппонент)
  ✨ Эффект: Сумеречное яйцо (#95). Сработал эффект карты Сумеречное яйцо.
  🃏 Розыгрыш: TLC_515 [TLC_515] 💎0
```

Пример JSON-фрагмента (`human`):

```json
{
  "mode": "human",
  "match_count": 1,
  "matches": [
    {
      "turns": [
        {
          "number": 1,
          "opening_hand": [],
          "events": [
            {
              "type": "effect_triggered",
              "summary": "Срабатывание эффекта"
            }
          ]
        }
      ]
    }
  ]
}
```

Пример JSON-фрагмента (`replay`):

```json
{
  "mode": "replay",
  "matches": [
    {
      "turns": [
        {
          "opening_state": {
            "players": {
              "PLAYER": {
                "mana": 1,
                "max_mana": 1,
                "hand_count": 3
              }
            }
          },
          "events": [
            {
              "type": "card_played",
              "action": {
                "player": "PLAYER"
              },
              "state_before": {},
              "state_after": {}
            }
          ]
        }
      ]
    }
  ]
}
```

## Известные ограничения

- `GameEndEvent` пока не восстанавливается из логов автоматически, поэтому победитель определяется только если событие будет добавлено отдельным парсером.
- Для части эффектов `effect_triggered` уже есть человекочитаемая расшифровка, но покрытие словаря `TriggerKeyword` еще не полное.
- Фильтры `filters.*` и `performance.workers` пока зарезервированы и не влияют на итоговый анализ.
- Автосохранение `timeline_*_{mode}.json` всегда пишет файл в каталог `output/` проекта.
- Текстовый вывод сейчас ориентирован на последний восстановленный матч, тогда как JSON уже отдает массив `matches`.

## Тестовые данные

Минимальные end-to-end fixtures лежат в `tests/testdata/`:

- `standard_game_1/logs` — компактный набор `LoadingScreen.log`, `Power.log`, `Zone.log`
- `cards/cards_sample.json` — минимальная база карт для e2e
- `expected/` — golden/expected данные для integration тестов

Запуск:

```bash
make test-integration
```

## Структура проекта

```text
hsanalyzer/
├── cmd/hsanalyzer/        # CLI на cobra
├── internal/analyzer/     # Analyzer и Pipeline
├── internal/carddb/       # База карт и update-cards
├── internal/config/       # Конфигурация
├── internal/formatter/    # Text / JSON human / JSON replay formatters
├── internal/logreader/    # Postrочное чтение логов
├── internal/parser/       # Парсеры логов
├── internal/state/        # Состояние игры
├── data/cards.json        # База карт
├── logs/                  # Примеры реальных логов
└── specs/                 # Design/tasks
```

## Разработка

Сборка:

```bash
make build
```

Все тесты:

```bash
go test ./...
```

Через `Makefile`:

```bash
make test
```

Полезные цели:

- `make build`
- `make run`
- `make test`
- `make test-unit`
- `make coverage`
- `make lint`

## Лицензия

MIT
