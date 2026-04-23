# Hearthstone Log Analyzer

Анализатор логов Hearthstone - инструмент для парсинга и анализа игровых логов Hearthstone с выводом детальной хронологии игры в человекочитаемом формате.

## Возможности

- 📊 Полный анализ игровых логов Hearthstone
- 🎮 Поддержка стандартного режима игры
- 📝 Детальная хронология всех событий игры
- 🃏 Отслеживание сыгранных карт, атак, эффектов
- 🤚 Мониторинг карт в руке игрока
- 🔍 Поддержка механики "Раскопки" (Discover)
- 🇷🇺 Локализация на русский язык
- ⚙️ Гибкая конфигурация

## Требования

- Go 1.21 или выше
- Hearthstone (для генерации логов)

## Установка

```bash
# Клонировать репозиторий
git clone https://github.com/yourusername/hsanalyzer.git
cd hsanalyzer

# Установить зависимости
go mod download

# Собрать проект
go build -o hsanalyzer.exe ./cmd/hsanalyzer
```

## Использование

### Базовое использование

```bash
# Анализ последней игры
./hsanalyzer analyze

# Указать путь к логам
./hsanalyzer analyze --logs-path "C:\Program Files (x86)\Hearthstone\Logs"

# Использовать конфигурационный файл
./hsanalyzer analyze --config config.yaml
```

### Конфигурация

Создайте файл `config.yaml`:

```yaml
logs_path: "C:\\Program Files (x86)\\Hearthstone\\Logs"
cards_db_path: './data/cards.json'

output:
  format: text
  destination: stdout
  language: ru
  verbose: false

filters:
  game_mode:
    - standard
  min_turns: 0
  max_turns: 0
```

## Структура проекта

```
hsanalyzer/
├── cmd/
│   └── hsanalyzer/          # CLI приложение
├── internal/
│   ├── analyzer/            # Оркестрация анализа
│   ├── config/              # Конфигурация
│   ├── parser/              # Парсеры логов
│   ├── game/                # Domain модели
│   ├── formatter/           # Форматтеры вывода
│   ├── carddb/              # База данных карт
│   └── logreader/           # Чтение логов
├── pkg/
│   └── hslog/               # Общие утилиты
├── data/
│   └── cards.json           # База карт
└── configs/
    └── config.yaml          # Пример конфигурации
```

## Архитектура

Проект построен на принципах Clean Architecture:

- **CLI Layer**: точка входа приложения
- **Application Layer**: оркестрация и pipeline обработки
- **Domain Layer**: бизнес-логика и модели игры
- **Infrastructure Layer**: парсеры, форматтеры, БД карт

## Разработка

### Запуск тестов

```bash
# Все тесты
make test

# Unit тесты
go test ./internal/...

# Integration тесты
go test ./tests/...
```

### Сборка

```bash
# Сборка для текущей платформы
make build

# Сборка для Windows
GOOS=windows GOARCH=amd64 go build -o hsanalyzer.exe ./cmd/hsanalyzer
```

## Пример вывода

```
=== Игра начата ===
Режим: Стандарт
Игрок: PlayerName (Маг)
Оппонент: OpponentName (Воин)

--- Ход 1 (Игрок) ---
💎 Мана: 1/1
🤚 Рука: [Огненный шар, Чародейский интеллект, Ледяная стрела]
📥 Взята карта: Огненный шар

--- Ход 2 (Оппонент) ---
💎 Мана: 2/2
🎴 Сыграна карта: Кровожадный берсерк (2💎)
  ⚔️ 2 / ❤️ 3

...
```

## Roadmap

- [x] Базовая структура проекта
- [x] Парсинг основных логов
- [x] Форматирование вывода
- [ ] Web интерфейс
- [ ] Статистика по играм
- [ ] Поддержка арены и полей сражений
- [ ] Real-time мониторинг

## Лицензия

MIT

## Контакты

- GitHub: [yourusername/hsanalyzer](https://github.com/yourusername/hsanalyzer)
- Issues: [github.com/yourusername/hsanalyzer/issues](https://github.com/yourusername/hsanalyzer/issues)
