# Hearthstone Log Analyzer

`hsanalyzer` — CLI-инструмент на Go для разбора логов Hearthstone, сборки игровой хронологии и вывода результата в текстовом виде.

## Текущий статус

Сейчас в проекте реализованы:

- чтение конфигурации через `yaml` и `viper`
- загрузка базы карт из `data/cards.json`
- парсеры для `Decks.log`, `Power.log`, `LoadingScreen.log` / `LoadingScreen_old.log`
- `Analyzer` и `Pipeline` для объединения, сортировки и применения событий
- текстовый formatter с поддержкой `ru` и `en`
- CLI на `cobra`
- structured logging через `zap`

Ограничения текущей версии:

- основной формат вывода только `text`
- `Zone.log` поддержан на уровне analyzer, но если файла нет, он просто пропускается
- поля `filters` и часть `performance` в конфиге пока зарезервированы и не влияют на итоговый анализ
- полный прогон `analyze` на реальном наборе логов может завершаться ошибкой на части `Power.log`, если для события не удалось восстановить нужную сущность в `GameState`

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
- `--format` — формат вывода, сейчас поддержан только `text`
- `--output` — `stdout` или `file`
- `--output-file` — путь к файлу результата, если `--output file`
- `--lang` — `ru` или `en`
- `--verbose` — расширенный текстовый вывод

Примеры:

```bash
go run ./cmd/hsanalyzer analyze --logs-path logs --cards-db-path data/cards.json

go run ./cmd/hsanalyzer analyze \
  --logs-path "C:\\Program Files (x86)\\Hearthstone\\Logs" \
  --cards-db-path data/cards.json \
  --lang ru \
  --verbose

go run ./cmd/hsanalyzer analyze \
  --config .\config.yaml \
  --output file \
  --output-file .\result.txt
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
logs_path: ./logs
cards_db_path: ./data/cards.json

output:
  format: text
  destination: stdout
  file_path: ""
  language: ru
  verbose: false

filters:
  game_mode:
    - standard
  min_turns: 0
  max_turns: 0

performance:
  workers: 1
  buffer_size: 65536
```

Что реально используется сейчас:

- `logs_path`
- `cards_db_path`
- `output.format`
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
- список карт колоды
- начало и конец ходов
- розыгрыш карт
- добор карт
- атаки
- урон
- смерти
- триггеры и изменения характеристик
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

## Структура проекта

```text
hsanalyzer/
├── cmd/hsanalyzer/        # CLI на cobra
├── internal/analyzer/     # Analyzer и Pipeline
├── internal/carddb/       # База карт и update-cards
├── internal/config/       # Конфигурация
├── internal/formatter/    # Formatter и TextFormatter
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

## Известные ограничения

- analyzer пока не покрывает все кейсы восстановления сущностей из `Power.log`
- formatter реализован только в текстовом виде
- JSON output и web/UI еще не реализованы
- интеграция `zap` есть только на уровне CLI

## Лицензия

MIT
