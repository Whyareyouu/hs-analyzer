# Implementation Plan

- [x] 1. Инициализация проекта и базовая структура

  - Создать структуру директорий согласно дизайну
  - Инициализировать Go модуль
  - Создать go.mod с необходимыми зависимостями
  - Создать базовый README.md с описанием проекта
  - _Requirements: 12.1, 12.2_

- [x] 2. Реализация domain моделей

  - [x] 2.1 Создать базовые типы и константы

    - Реализовать типы PlayerType, ZoneType, CardType, EventType в internal/game
    - Создать константы для игровых режимов в game_mode.go
    - Написать unit тесты для типов
    - _Requirements: 10.1, 10.3_

  - [x] 2.2 Реализовать структуры карт и сущностей

    - Создать Card структуру в internal/game/card.go
    - Создать Entity структуру с EntityAttributes в internal/game/entity.go
    - Создать Zone структуру в internal/game/zone.go
    - Написать unit тесты для структур
    - _Requirements: 3.3, 3.4, 5.1_

  - [x] 2.3 Реализовать Player и GameState

    - Создать Player структуру с зонами в internal/game/player.go
    - Создать GameState структуру в internal/game/state.go
    - Реализовать методы для управления состоянием
    - Написать unit тесты для Player и GameState
    - _Requirements: 2.1, 2.2, 5.2, 5.3_

  - [x] 2.4 Создать Turn структуру

    - Реализовать Turn в internal/game/turn.go
    - Добавить методы для добавления событий в ход
    - Написать unit тесты для Turn
    - _Requirements: 2.3, 2.4, 2.5_

- [x] 3. Реализация Event интерфейса и типов событий

  - [x] 3.1 Создать базовый Event интерфейс

    - Определить Event интерфейс в internal/game/event.go
    - Создать BaseEvent структуру с общими полями
    - Реализовать EventType константы
    - _Requirements: 2.1, 2.3_

  - [x] 3.2 Реализовать события игры и ходов

    - Создать GameStartEvent и GameEndEvent
    - Создать TurnStartEvent и TurnEndEvent
    - Реализовать метод Apply для каждого события
    - Написать unit тесты
    - _Requirements: 1.6, 2.3, 2.4, 8.7_

  - [x] 3.3 Реализовать события карт

    - Создать CardPlayedEvent в internal/game/event.go
    - Создать CardDrawnEvent
    - Реализовать методы Apply для обновления GameState
    - Написать unit тесты
    - _Requirements: 3.1, 3.2, 3.6, 5.2, 5.3_

  - [x] 3.4 Реализовать события боя

    - Создать AttackEvent
    - Создать DamageEvent и HealEvent
    - Создать DeathEvent
    - Реализовать методы Apply
    - Написать unit тесты
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.6_

  - [x] 3.5 Реализовать события эффектов
    - Создать EffectTriggeredEvent
    - Создать DiscoverEvent
    - Реализовать методы Apply
    - Написать unit тесты
    - _Requirements: 6.1, 6.2, 6.7, 7.1, 7.2, 7.4_

- [x] 4. Реализация CardDatabase

  - [x] 4.1 Создать интерфейс и структуры CardInfo

    - Определить CardDatabase интерфейс в internal/carddb/database.go
    - Создать CardInfo структуру в internal/carddb/card_info.go
    - Добавить типы CardType, Rarity, CardClass, Race
    - _Requirements: 9.3_

  - [x] 4.2 Реализовать JSON загрузчик базы карт

    - Создать JSONDatabase в internal/carddb/json_database.go
    - Реализовать метод Load для загрузки из JSON
    - Реализовать методы GetCard и GetCardByName с кешированием
    - Написать unit тесты с mock данными
    - _Requirements: 9.1, 9.2, 9.4_

  - [x] 4.3 Добавить поддержку локализации
    - Добавить поля NameRu и TextRu в CardInfo
    - Реализовать метод GetLocalizedName
    - Написать unit тесты
    - _Requirements: 9.5, 8.3_

- [x] 5. Реализация конфигурации

  - [x] 5.1 Создать структуры конфигурации

    - Создать Config, OutputConfig, FilterConfig в internal/config/config.go
    - Добавить yaml теги для всех полей
    - _Requirements: 12.1, 12.5_

  - [x] 5.2 Реализовать загрузчик конфигурации

    - Создать Loader в internal/config/loader.go
    - Интегрировать viper для чтения YAML
    - Реализовать валидацию путей к файлам
    - Написать unit тесты
    - _Requirements: 12.2, 12.4, 11.5_

  - [x] 5.3 Добавить значения по умолчанию
    - Создать defaults.go с дефолтными значениями
    - На первой версии приложения, все необходимые логи всегда будут находится в logs
    - Написать unit тесты
    - _Requirements: 12.3_

- [x] 6. Реализация LogReader

  - [x] 6.1 Создать интерфейс LogReader

    - Определить LogReader интерфейс в internal/logreader/reader.go
    - Добавить методы для чтения логов построчно
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 6.2 Реализовать FileReader
    - Создать FileReader в internal/logreader/file_reader.go
    - Использовать bufio.Scanner для эффективного чтения
    - Добавить обработку ошибок чтения файлов
    - Написать unit тесты
    - _Requirements: 1.5, 11.1_

- [x] 7. Реализация парсеров логов

  - [x] 7.1 Создать базовый Parser интерфейс и утилиты

    - Определить Parser интерфейс в internal/parser/parser.go
    - Создать LogType константы
    - Создать общие regex паттерны в pkg/hslog/patterns.go
    - Создать константы из логов HS в pkg/hslog/constants.go
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 7.2 Реализовать LoadingScreenParser

    - Создать LoadingScreenParser в internal/parser/loading_parser.go
    - Парсить начало игры и режим игры
    - Генерировать GameStartEvent
    - Написать unit тесты с примерами логов
    - _Requirements: 1.3, 1.6, 10.1, 10.2_

  - [x] 7.3 Реализовать DeckParser

    - Создать DeckParser в internal/parser/deck_parser.go
    - Парсить список карт в колоде
    - Связывать карты с игроком
    - Написать unit тесты
    - _Requirements: 1.4, 5.1_

  - [x] 7.4 Реализовать ZoneParser

    - Создать ZoneParser в internal/parser/zone_parser.go
    - Парсить перемещения карт между зонами
    - Генерировать CardDrawnEvent
    - Написать unit тесты
    - _Requirements: 1.2, 5.2, 5.4_

  - [x] 7.5 Реализовать PowerParser - базовая структура

    - Создать PowerParser в internal/parser/power_parser.go
    - Реализовать базовую структуру парсера
    - Добавить regex паттерны для основных событий
    - _Requirements: 1.1_

  - [x] 7.6 PowerParser - парсинг ходов

    - Добавить парсинг TAG_CHANGE для определения ходов
    - Генерировать TurnStartEvent и TurnEndEvent
    - Отслеживать текущего игрока
    - Написать unit тесты
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 7.7 PowerParser - парсинг розыгрыша карт

    - Добавить парсинг ACTION_START для розыгрыша карт
    - Парсить FULL_ENTITY для новых сущностей
    - Генерировать CardPlayedEvent
    - Написать unit тесты
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.6_

  - [x] 7.8 PowerParser - парсинг атак

    - Добавить парсинг ATTACK событий
    - Генерировать AttackEvent
    - Отслеживать урон и смерти
    - Написать unit тесты
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.6_

  - [x] 7.9 PowerParser - парсинг эффектов

    - Добавить парсинг TRIGGER и POWER событий
    - Генерировать EffectTriggeredEvent
    - Парсить изменения характеристик
    - Написать unit тесты
    - _Requirements: 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.7, 4.5_

  - [x] 7.10 PowerParser - парсинг Discover

    - Добавить парсинг SHOW_ENTITY для Discover
    - Генерировать DiscoverEvent с опциями
    - Отслеживать выбранную карту
    - Написать unit тесты
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 5.6_

  - [x] 7.11 Создать ParserRegistry
    - Создать Registry в internal/parser/registry.go
    - Реализовать регистрацию и получение парсеров
    - Написать unit тесты
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 8. Реализация Analyzer и Pipeline

  - [x] 8.1 Создать Pipeline для обработки событий

    - Создать Pipeline в internal/analyzer/pipeline.go
    - Определить EventHandler интерфейс
    - Реализовать цепочку обработчиков с приоритетами
    - Написать unit тесты
    - _Requirements: 2.5, 11.4_

  - [ ] 8.2 Создать основной Analyzer

    - Создать Analyzer в internal/analyzer/analyzer.go
    - Реализовать dependency injection для всех компонентов
    - Добавить метод Analyze для оркестрации процесса
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ] 8.3 Реализовать чтение и парсинг логов

    - Добавить логику чтения всех типов логов
    - Вызывать соответствующие парсеры
    - Собирать события из всех логов
    - Обрабатывать ошибки gracefully
    - Написать integration тесты
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 11.1, 11.4_

  - [ ] 8.4 Реализовать объединение и сортировку событий

    - Объединить события из разных логов
    - Сортировать по timestamp
    - Определить границы игровой сессии
    - Написать unit тесты
    - _Requirements: 1.6, 1.7_

  - [ ] 8.5 Реализовать применение событий к состоянию
    - Создать начальное GameState
    - Применить события через Pipeline
    - Обновлять состояние игры
    - Обрабатывать ошибки
    - Написать integration тесты
    - _Requirements: 2.5, 11.2, 11.3, 11.4_

- [ ] 9. Реализация Formatter

  - [ ] 9.1 Создать Formatter интерфейс

    - Определить Formatter интерфейс в internal/formatter/formatter.go
    - Добавить методы Format и FormatEvent
    - _Requirements: 8.1_

  - [ ] 9.2 Реализовать TextFormatter - базовая структура

    - Создать TextFormatter в internal/formatter/text_formatter.go
    - Добавить конфигурацию для языка и verbose режима
    - _Requirements: 8.2, 8.3, 8.5_

  - [ ] 9.3 TextFormatter - форматирование ходов

    - Реализовать форматирование начала/конца хода
    - Добавить отображение номера хода и игрока
    - Использовать отступы для структурирования
    - Написать unit тесты
    - _Requirements: 8.2, 8.5_

  - [ ] 9.4 TextFormatter - форматирование карт

    - Реализовать форматирование розыгрыша карт
    - Добавить отображение характеристик карт
    - Использовать эмодзи для визуализации (⚔️❤️💎)
    - Написать unit тесты
    - _Requirements: 8.3, 8.4, 8.6_

  - [ ] 9.5 TextFormatter - форматирование боя

    - Реализовать форматирование атак
    - Добавить отображение урона и смертей
    - Написать unit тесты
    - _Requirements: 8.4_

  - [ ] 9.6 TextFormatter - форматирование эффектов

    - Реализовать форматирование эффектов карт
    - Добавить отображение Discover опций
    - Написать unit тесты
    - _Requirements: 8.3_

  - [ ] 9.7 TextFormatter - форматирование состояния руки

    - Реализовать отображение карт в руке игрока
    - Добавить группировку по ходам
    - Написать unit тесты
    - _Requirements: 8.4, 8.5_

  - [ ] 9.8 TextFormatter - итоговый вывод

    - Реализовать форматирование результата игры
    - Добавить общую статистику
    - Написать unit тесты
    - _Requirements: 8.7_

  - [ ] 9.9 Создать FormatterRegistry
    - Создать Registry в internal/formatter/registry.go
    - Зарегистрировать TextFormatter
    - Написать unit тесты
    - _Requirements: 8.1_

- [ ] 10. Реализация CLI

  - [ ] 10.1 Создать основную структуру CLI

    - Создать main.go в cmd/hsanalyzer
    - Интегрировать cobra для CLI
    - Добавить root command
    - _Requirements: 12.1_

  - [ ] 10.2 Добавить команду analyze

    - Создать analyze command
    - Добавить флаги для путей к логам и конфигурации
    - Реализовать инициализацию всех компонентов
    - _Requirements: 12.2, 12.4, 12.5_

  - [ ] 10.3 Интегрировать все компоненты

    - Создать dependency injection контейнер
    - Инициализировать Config, CardDB, Parsers, Analyzer, Formatter
    - Запустить анализ и вывести результат
    - Обработать ошибки с понятными сообщениями
    - _Requirements: 11.5, 12.1, 12.5_

  - [ ] 10.4 Добавить логирование
    - Интегрировать zap logger
    - Добавить структурированное логирование
    - Настроить уровни логирования
    - _Requirements: 11.3_

- [ ] 11. Тестирование и документация

  - [ ] 11.1 Создать тестовые данные

    - Создать директорию tests/testdata
    - Добавить примеры логов стандартных игр
    - Создать sample cards.json
    - Создать ожидаемые выходные данные
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ] 11.2 Написать integration тесты

    - Создать end-to-end тесты с реальными логами
    - Тестировать полный pipeline от логов до вывода
    - Проверить различные сценарии игр
    - _Requirements: 1.7, 8.1, 8.7_

  - [ ] 11.3 Создать Makefile

    - Добавить команды для сборки
    - Добавить команды для запуска тестов
    - Добавить команды для линтинга
    - _Requirements: 12.1_

  - [ ] 11.4 Написать README.md

    - Добавить описание проекта
    - Добавить инструкции по установке
    - Добавить примеры использования
    - Добавить информацию о конфигурации
    - _Requirements: 12.1, 12.2, 12.3, 12.4_

  - [ ] 11.5 Создать пример конфигурации
    - Создать configs/config.yaml
    - Добавить комментарии для всех опций
    - Указать значения по умолчанию
    - _Requirements: 12.1, 12.3_

- [ ] 12. Финальная интеграция и тестирование

  - [ ] 12.1 Провести полное тестирование

    - Запустить все unit тесты
    - Запустить integration тесты
    - Протестировать с реальными логами Hearthstone
    - Проверить обработку ошибок
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [ ] 12.2 Оптимизация производительности

    - Профилировать приложение
    - Оптимизировать узкие места
    - Проверить использование памяти
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ] 12.3 Финальная документация
    - Обновить README с актуальной информацией
    - Добавить примеры вывода
    - Документировать известные ограничения
    - _Requirements: 10.1, 10.2, 10.3_
