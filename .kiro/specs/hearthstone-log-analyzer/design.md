# Design Document: Hearthstone Log Analyzer

## Overview

Hearthstone Log Analyzer - это приложение на Go для парсинга и анализа игровых логов Hearthstone. Архитектура построена на принципах Clean Architecture с четким разделением ответственности, использованием интерфейсов для расширяемости и следованием best practices Go.

Ключевые принципы дизайна:

- **Модульность**: каждый компонент имеет четкую ответственность
- **Расширяемость**: легко добавлять новые парсеры, форматтеры и обработчики событий
- **Тестируемость**: все зависимости через интерфейсы
- **Производительность**: эффективная обработка больших логов
- **Maintainability**: понятная структура кода, следование Go conventions

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│                    (cmd/hsanalyzer)                          │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Application Layer                         │
│                  (internal/analyzer)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Analyzer   │  │ Orchestrator │  │   Pipeline   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                     Domain Layer                             │
│                   (internal/game)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  State   │  │  Event   │  │   Card   │  │   Turn   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │     Parsers      │  │    Formatters    │                │
│  │  (internal/      │  │  (internal/      │                │
│  │   parser)        │  │   formatter)     │                │
│  └──────────────────┘  └──────────────────┘                │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │    Card DB       │  │     Config       │                │
│  │  (internal/      │  │  (internal/      │                │
│  │   carddb)        │  │   config)        │                │
│  └──────────────────┘  └──────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
hsanalyzer/
├── cmd/
│   └── hsanalyzer/
│       └── main.go                 # Точка входа, инициализация DI
│
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go            # Главный оркестратор анализа
│   │   ├── pipeline.go            # Pipeline для обработки событий
│   │   └── options.go             # Опции конфигурации анализатора
│   │
│   ├── config/
│   │   ├── config.go              # Структура конфигурации
│   │   ├── loader.go              # Загрузка конфигурации
│   │   └── defaults.go            # Значения по умолчанию
│   │
│   ├── parser/
│   │   ├── parser.go              # Интерфейс Parser
│   │   ├── power_parser.go        # Парсер Power.log
│   │   ├── zone_parser.go         # Парсер Zone.log
│   │   ├── loading_parser.go      # Парсер LoadingScreen.log
│   │   ├── deck_parser.go         # Парсер Decks.log
│   │   ├── registry.go            # Регистрация парсеров
│   │   └── common.go              # Общие утилиты парсинга
│   │
│   ├── game/
│   │   ├── state.go               # GameState - состояние игры
│   │   ├── event.go               # Event интерфейс и типы событий
│   │   ├── card.go                # Card структура
│   │   ├── entity.go              # Entity (миньоны, герои)
│   │   ├── turn.go                # Turn структура
│   │   ├── player.go              # Player структура
│   │   ├── zone.go                # Zone (рука, доска, колода)
│   │   ├── action.go              # Action (атаки, эффекты)
│   │   └── game_mode.go           # GameMode (стандарт, арена и т.д.)
│   │
│   ├── formatter/
│   │   ├── formatter.go           # Интерфейс Formatter
│   │   ├── text_formatter.go      # Текстовый форматтер
│   │   ├── json_formatter.go      # JSON форматтер (для будущего)
│   │   └── registry.go            # Регистрация форматтеров
│   │
│   ├── carddb/
│   │   ├── database.go            # Интерфейс CardDatabase
│   │   ├── json_database.go       # JSON реализация
│   │   ├── card_info.go           # CardInfo структура
│   │   └── loader.go              # Загрузка базы карт
│   │
│   └── logreader/
│       ├── reader.go              # Интерфейс LogReader
│       ├── file_reader.go         # Чтение из файлов
│       └── buffer.go              # Буферизация для больших файлов
│
├── pkg/
│   └── hslog/
│       ├── patterns.go            # Regex паттерны для логов
│       └── constants.go           # Константы из логов HS
│
├── data/
│   └── cards.json                 # База карт Hearthstone
│
├── configs/
│   └── config.yaml                # Пример конфигурации
│
├── go.mod
├── go.sum
├── Makefile                       # Команды для сборки и тестирования
└── README.md
```

## Components and Interfaces

### 1. Core Interfaces

#### Parser Interface

```go
// Parser определяет интерфейс для парсинга различных типов логов
type Parser interface {
    // Parse читает лог и возвращает события
    Parse(ctx context.Context, reader io.Reader) ([]game.Event, error)

    // Name возвращает имя парсера
    Name() string

    // LogType возвращает тип лога, который парсит этот парсер
    LogType() LogType
}

// LogType определяет тип лога
type LogType string

const (
    LogTypePower        LogType = "Power"
    LogTypeZone         LogType = "Zone"
    LogTypeLoadingScreen LogType = "LoadingScreen"
    LogTypeDeck         LogType = "Deck"
)
```

#### Event Interface

```go
// Event представляет игровое событие
type Event interface {
    // Type возвращает тип события
    Type() EventType

    // Timestamp возвращает время события
    Timestamp() time.Time

    // TurnNumber возвращает номер хода
    TurnNumber() int

    // Apply применяет событие к состоянию игры
    Apply(state *GameState) error
}

// EventType определяет тип события
type EventType string

const (
    EventTypeGameStart      EventType = "game_start"
    EventTypeTurnStart      EventType = "turn_start"
    EventTypeTurnEnd        EventType = "turn_end"
    EventTypeCardPlayed     EventType = "card_played"
    EventTypeCardDrawn      EventType = "card_drawn"
    EventTypeAttack         EventType = "attack"
    EventTypeDamage         EventType = "damage"
    EventTypeHeal           EventType = "heal"
    EventTypeDeath          EventType = "death"
    EventTypeEffectTriggered EventType = "effect_triggered"
    EventTypeDiscover       EventType = "discover"
    EventTypeGameEnd        EventType = "game_end"
)
```

#### Formatter Interface

```go
// Formatter определяет интерфейс для форматирования вывода
type Formatter interface {
    // Format форматирует игровое состояние в строку
    Format(game *Game) (string, error)

    // FormatEvent форматирует отдельное событие
    FormatEvent(event Event) (string, error)

    // Name возвращает имя форматтера
    Name() string
}
```

#### CardDatabase Interface

```go
// CardDatabase определяет интерфейс для работы с базой карт
type CardDatabase interface {
    // GetCard возвращает информацию о карте по ID
    GetCard(cardID string) (*CardInfo, error)

    // GetCardByName возвращает карту по имени
    GetCardByName(name string) (*CardInfo, error)

    // Load загружает базу данных
    Load(ctx context.Context) error

    // IsLoaded проверяет, загружена ли база
    IsLoaded() bool
}
```

### 2. Domain Models

#### GameState

```go
// GameState представляет полное состояние игры
type GameState struct {
    GameID          string
    GameMode        GameMode
    CurrentTurn     int
    CurrentPlayer   PlayerType

    Player          *Player
    Opponent        *Player

    Turns           []*Turn
    Events          []Event

    StartTime       time.Time
    EndTime         *time.Time
    Winner          *PlayerType

    metadata        map[string]interface{}
}

// Player представляет игрока
type Player struct {
    ID              int
    Name            string
    Type            PlayerType

    Hero            *Entity
    HeroPower       *Card
    Weapon          *Card

    Hand            *Zone
    Deck            *Zone
    Board           *Zone
    Graveyard       *Zone

    Mana            int
    MaxMana         int
    OverloadMana    int

    DeckList        []*Card  // Изначальный список колоды
}

// Entity представляет сущность на доске (миньон, герой)
type Entity struct {
    ID              int
    CardID          string
    Card            *CardInfo

    Attack          int
    Health          int
    MaxHealth       int

    Position        int
    Owner           PlayerType

    Attributes      EntityAttributes
    Enchantments    []*Enchantment

    CreatedBy       *Entity
    CreatedByEffect string
}

// EntityAttributes содержит атрибуты сущности
type EntityAttributes struct {
    Taunt           bool
    DivineShield    bool
    Stealth         bool
    Windfury        bool
    Charge          bool
    Rush            bool
    Lifesteal       bool
    Poisonous       bool
    Frozen          bool
    Silenced        bool
    Immune          bool
}

// Turn представляет игровой ход
type Turn struct {
    Number          int
    Player          PlayerType
    Events          []Event
    StartTime       time.Time
    EndTime         *time.Time

    CardsPlayed     []*Card
    AttacksMade     []*Attack
    CardsDrawn      []*Card
}

// Card представляет карту
type Card struct {
    ID              string
    EntityID        int
    Info            *CardInfo

    Cost            int
    OriginalCost    int

    Owner           PlayerType
    Controller      PlayerType

    Zone            ZoneType
    Position        int

    CreatedBy       string
}

// Zone представляет игровую зону
type Zone struct {
    Type            ZoneType
    Cards           []*Card
    Owner           PlayerType
}

// ZoneType определяет тип зоны
type ZoneType string

const (
    ZoneHand        ZoneType = "HAND"
    ZoneDeck        ZoneType = "DECK"
    ZonePlay        ZoneType = "PLAY"
    ZoneGraveyard   ZoneType = "GRAVEYARD"
    ZoneSecret      ZoneType = "SECRET"
    ZoneSetAside    ZoneType = "SETASIDE"
)
```

#### Event Types

```go
// CardPlayedEvent - событие розыгрыша карты
type CardPlayedEvent struct {
    BaseEvent
    Card            *Card
    Player          PlayerType
    Target          *Entity
    Position        int
}

// AttackEvent - событие атаки
type AttackEvent struct {
    BaseEvent
    Attacker        *Entity
    Defender        *Entity
    Damage          int
}

// DiscoverEvent - событие раскопок
type DiscoverEvent struct {
    BaseEvent
    Player          PlayerType
    Options         []*CardInfo
    Chosen          *CardInfo
    Source          string
}

// EffectTriggeredEvent - событие срабатывания эффекта
type EffectTriggeredEvent struct {
    BaseEvent
    Source          *Entity
    EffectName      string
    Targets         []*Entity
    Values          map[string]interface{}
}
```

### 3. Application Layer

#### Analyzer

```go
// Analyzer - главный компонент анализа
type Analyzer struct {
    config          *config.Config
    cardDB          carddb.CardDatabase
    parsers         map[parser.LogType]parser.Parser
    formatter       formatter.Formatter
    pipeline        *Pipeline
}

// Analyze выполняет полный анализ игры
func (a *Analyzer) Analyze(ctx context.Context) (*game.Game, error) {
    // 1. Читаем все логи
    // 2. Парсим события из каждого лога
    // 3. Объединяем и сортируем события
    // 4. Применяем события к состоянию игры
    // 5. Возвращаем финальное состояние
}

// Pipeline обрабатывает события через цепочку обработчиков
type Pipeline struct {
    handlers        []EventHandler
}

// EventHandler обрабатывает события
type EventHandler interface {
    Handle(ctx context.Context, event game.Event, state *game.GameState) error
    Priority() int
}
```

## Data Models

### CardInfo Structure

```go
// CardInfo содержит информацию о карте из базы данных
type CardInfo struct {
    ID              string            `json:"id"`
    DBFId           int               `json:"dbfId"`
    Name            string            `json:"name"`
    NameRu          string            `json:"nameRu"`
    Text            string            `json:"text"`
    TextRu          string            `json:"textRu"`

    Type            CardType          `json:"type"`
    Rarity          Rarity            `json:"rarity"`
    Class           CardClass         `json:"class"`
    Set             CardSet           `json:"set"`

    Cost            int               `json:"cost"`
    Attack          *int              `json:"attack,omitempty"`
    Health          *int              `json:"health,omitempty"`
    Durability      *int              `json:"durability,omitempty"`

    Mechanics       []string          `json:"mechanics,omitempty"`
    Race            *Race             `json:"race,omitempty"`

    Collectible     bool              `json:"collectible"`
    Elite           bool              `json:"elite"`

    Flavor          string            `json:"flavor,omitempty"`
    Artist          string            `json:"artist,omitempty"`
}

// CardType определяет тип карты
type CardType string

const (
    CardTypeMinion      CardType = "MINION"
    CardTypeSpell       CardType = "SPELL"
    CardTypeWeapon      CardType = "WEAPON"
    CardTypeHero        CardType = "HERO"
    CardTypeHeroPower   CardType = "HERO_POWER"
)
```

### Configuration Model

```go
// Config содержит конфигурацию приложения
type Config struct {
    // Пути к логам
    LogsPath        string            `yaml:"logs_path"`

    // База данных карт
    CardsDBPath     string            `yaml:"cards_db_path"`

    // Настройки вывода
    Output          OutputConfig      `yaml:"output"`

    // Фильтры
    Filters         FilterConfig      `yaml:"filters"`

    // Производительность
    Performance     PerformanceConfig `yaml:"performance"`
}

// OutputConfig настройки вывода
type OutputConfig struct {
    Format          string            `yaml:"format"`  // text, json
    Destination     string            `yaml:"destination"` // stdout, file
    FilePath        string            `yaml:"file_path,omitempty"`
    Language        string            `yaml:"language"` // ru, en
    Verbose         bool              `yaml:"verbose"`
}

// FilterConfig фильтры для обработки
type FilterConfig struct {
    GameMode        []string          `yaml:"game_mode"`
    MinTurns        int               `yaml:"min_turns"`
    MaxTurns        int               `yaml:"max_turns"`
}
```

## Error Handling

### Error Types

```go
// Определяем специфичные типы ошибок
var (
    ErrLogNotFound          = errors.New("log file not found")
    ErrInvalidLogFormat     = errors.New("invalid log format")
    ErrCardNotFound         = errors.New("card not found in database")
    ErrInvalidGameMode      = errors.New("invalid game mode")
    ErrParsingFailed        = errors.New("parsing failed")
    ErrDatabaseNotLoaded    = errors.New("card database not loaded")
)

// AppError содержит контекст ошибки
type AppError struct {
    Op              string          // Операция, где произошла ошибка
    Err             error           // Оригинальная ошибка
    Context         map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *AppError) Unwrap() error {
    return e.Err
}
```

### Error Handling Strategy

1. **Graceful Degradation**: при ошибке парсинга одного события продолжаем обработку
2. **Logging**: все ошибки логируются с контекстом
3. **User-Friendly Messages**: понятные сообщения для пользователя
4. **Recovery**: попытка восстановления при частичных ошибках

## Testing Strategy

### Unit Tests

- Тесты для каждого парсера с mock данными
- Тесты для domain моделей и их методов
- Тесты для форматтеров
- Тесты для CardDatabase

### Integration Tests

- Тесты полного pipeline с реальными логами
- Тесты взаимодействия компонентов
- Тесты различных сценариев игр

### Test Data

```
tests/
├── testdata/
│   ├── logs/
│   │   ├── standard_game_1/
│   │   │   ├── Power.log
│   │   │   ├── Zone.log
│   │   │   └── LoadingScreen.log
│   │   └── standard_game_2/
│   ├── cards/
│   │   └── cards_sample.json
│   └── expected/
│       ├── game_1_output.txt
│       └── game_2_output.txt
└── mocks/
    ├── mock_parser.go
    ├── mock_carddb.go
    └── mock_formatter.go
```

### Testing Approach

```go
// Пример теста парсера
func TestPowerParser_Parse(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        want        []game.Event
        wantErr     bool
    }{
        {
            name: "parse card played event",
            input: "D 12:34:56.789 GameState.DebugPrintPower() - ACTION_START...",
            want: []game.Event{
                &game.CardPlayedEvent{...},
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := parser.NewPowerParser()
            got, err := p.Parse(context.Background(), strings.NewReader(tt.input))

            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Performance Considerations

### Optimization Strategies

1. **Streaming Processing**

   - Читаем логи построчно, не загружая весь файл в память
   - Используем `bufio.Scanner` для эффективного чтения

2. **Concurrent Parsing**

   - Парсим разные логи параллельно с использованием goroutines
   - Используем worker pool для обработки событий

3. **Memory Management**

   - Переиспользуем объекты где возможно (sync.Pool)
   - Избегаем лишних аллокаций
   - Используем string interning для повторяющихся строк

4. **Caching**
   - Кешируем результаты парсинга regex
   - Кешируем CardInfo из базы данных

```go
// Пример использования worker pool
type EventProcessor struct {
    workers     int
    eventChan   chan game.Event
    resultChan  chan *ProcessedEvent
    wg          sync.WaitGroup
}

func (p *EventProcessor) Process(ctx context.Context, events []game.Event) error {
    // Запускаем workers
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }

    // Отправляем события на обработку
    go func() {
        for _, event := range events {
            select {
            case p.eventChan <- event:
            case <-ctx.Done():
                return
            }
        }
        close(p.eventChan)
    }()

    // Ждем завершения
    p.wg.Wait()
    close(p.resultChan)

    return nil
}
```

## Extensibility Points

### 1. Adding New Parsers

```go
// Регистрация нового парсера
func init() {
    parser.Register(parser.LogTypeCustom, NewCustomParser())
}

type CustomParser struct {
    // ...
}

func (p *CustomParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
    // Реализация парсинга
}
```

### 2. Adding New Event Types

```go
// Новый тип события
type CustomEvent struct {
    BaseEvent
    CustomField string
}

func (e *CustomEvent) Type() EventType {
    return "custom_event"
}

func (e *CustomEvent) Apply(state *GameState) error {
    // Применение к состоянию
}
```

### 3. Adding New Formatters

```go
// Регистрация нового форматтера
func init() {
    formatter.Register("html", NewHTMLFormatter())
}

type HTMLFormatter struct {
    // ...
}

func (f *HTMLFormatter) Format(game *Game) (string, error) {
    // Генерация HTML
}
```

### 4. Adding Event Handlers

```go
// Новый обработчик событий
type StatisticsHandler struct {
    stats *Statistics
}

func (h *StatisticsHandler) Handle(ctx context.Context, event game.Event, state *game.GameState) error {
    // Сбор статистики
    switch e := event.(type) {
    case *game.CardPlayedEvent:
        h.stats.CardsPlayed++
    case *game.AttackEvent:
        h.stats.AttacksMade++
    }
    return nil
}

func (h *StatisticsHandler) Priority() int {
    return 100 // Низкий приоритет, выполняется после основных обработчиков
}
```

## Dependencies

### External Libraries

```go
// go.mod
module github.com/username/hsanalyzer

go 1.21

require (
    gopkg.in/yaml.v3 v3.0.1              // Конфигурация
    github.com/spf13/cobra v1.8.0         // CLI
    github.com/spf13/viper v1.18.2        // Конфигурация
    go.uber.org/zap v1.26.0               // Логирование
    github.com/stretchr/testify v1.8.4    // Тестирование
)
```

### Rationale for Dependencies

- **cobra**: стандарт для CLI приложений в Go
- **viper**: гибкая работа с конфигурацией
- **zap**: высокопроизводительное структурированное логирование
- **testify**: удобные assertion'ы для тестов

## Implementation Phases

### Phase 1: Core Infrastructure (MVP)

- Базовая структура проекта
- Конфигурация и загрузка настроек
- CardDatabase с загрузкой из JSON
- Базовые domain модели (GameState, Player, Card, Entity)

### Phase 2: Log Parsing

- PowerParser для основных событий
- ZoneParser для отслеживания зон
- LoadingScreenParser для определения начала игры
- Объединение событий из разных логов

### Phase 3: Game State Management

- Применение событий к состоянию
- Отслеживание ходов
- Управление зонами (рука, доска, колода)
- Обработка эффектов

### Phase 4: Output Formatting

- TextFormatter с человекочитаемым выводом
- Форматирование по ходам
- Отображение состояния доски
- Локализация на русский язык

### Phase 5: Polish & Testing

- Comprehensive тесты
- Обработка edge cases
- Оптимизация производительности
- Документация

## Security Considerations

1. **File Access**: валидация путей к файлам, предотвращение path traversal
2. **Input Validation**: проверка формата логов перед парсингом
3. **Resource Limits**: ограничение размера обрабатываемых файлов
4. **Error Information**: не раскрываем системную информацию в ошибках

## Monitoring and Logging

```go
// Структурированное логирование
logger.Info("parsing log file",
    zap.String("file", logPath),
    zap.String("type", string(logType)),
    zap.Int("events_found", len(events)),
)

logger.Error("failed to parse event",
    zap.Error(err),
    zap.String("line", line),
    zap.Int("line_number", lineNum),
)

// Метрики (для будущего расширения)
type Metrics struct {
    ParseDuration       time.Duration
    EventsProcessed     int
    ErrorsEncountered   int
    MemoryUsed          uint64
}
```

## Future Enhancements

1. **Web Interface**: веб-интерфейс для просмотра игр
2. **Database Storage**: сохранение игр в БД для истории
3. **Statistics**: детальная статистика по играм
4. **Replay System**: воспроизведение игр
5. **Arena/Battlegrounds Support**: поддержка других режимов
6. **Real-time Monitoring**: отслеживание игры в реальном времени
7. **Deck Analysis**: анализ эффективности колод
8. **Export Formats**: экспорт в различные форматы (JSON, CSV, HTML)
