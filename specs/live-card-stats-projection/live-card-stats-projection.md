# Live Card Stats Projection (Task 13)

## Цель

Определить следующий этап live-monitoring режима: near-real-time проекцию сыгранных карт для внешнего клиента.

Проекция должна отвечать на простой продуктовый вопрос:

- какие карты уже сыграл я
- какие карты уже сыграл оппонент
- сколько раз каждая карта была сыграна

Для клиента по каждой карте нужен минимальный payload:

- `name`
- `description`
- `image`

## Правило приемки

Каждый пункт этой спецификации считается принятым только тогда, когда для него:

- определены соответствующие тесты
- тесты реализованы
- тесты проходят

Ни один пункт не считается завершенным только по факту описания или реализации без подтверждающих тестов.

## 1. Product Scope

Новая функциональность покрывает только активный матч в live runtime:

- накопление статистики сыгранных карт по двум сторонам: `self` и `opponent`
- публикацию двух отдельных массивов сыгранных карт: одного для `self`, другого для `opponent`
- обновление данных сразу после получения live `CardPlayedEvent`
- enrichment карточных данных через локальную card database
- публикацию lightweight snapshot для клиента без полной materialization replay/timeline

Функциональность не покрывает на этом этапе:

- скрытые карты в руке или колоде
- долгосрочную cross-match статистику
- полный turn-by-turn replay для клиента
- board state, mana curve, lethal hints и другие overlay-фичи
- жесткую привязку к конкретному transport слою вроде WebSocket или HTTP API

## 2. Source Of Truth

Источником факта "карта сыграна" является live поток `CardPlayedEvent`, уже порождаемый parser/pipeline в monitor flow.

Используемые поля:

- сторона определяется по `CardPlayedEvent.Player`
- идентичность карты определяется по `CardPlayedEvent.Card.ID`
- карточные метаданные резолвятся из `data/cards.json`

Важно:

- проекция не должна ждать `GameEnd`
- проекция не должна строиться через повторную materialization всего матча
- проекция обновляется инкрементально по входящим live events

## 3. Visibility Policy

Проекция предназначена для клиентского live view и поэтому должна содержать только публично наблюдаемую информацию.

Разрешено:

- карты, которые реально были сыграны игроком
- карты, которые реально были сыграны оппонентом

Запрещено:

- карты, только находящиеся в руке
- карты, только замеченные в decklist, если они не были сыграны
- догадки, инференс архетипа или скрытая информация о колоде оппонента

Это означает, что opponent bucket является strictly public projection: туда попадают только фактически разыгранные карты.

## 4. Aggregation Rules

Статистика агрегируется по `card_id`, а не по `entity_id`.

Следствия:

- если одна и та же карта сыграна несколько раз одной стороной, запись остается одна, а `play_count` увеличивается
- копии одной и той же карты, созданные эффектами, тоже агрегируются в тот же `card_id`, если их `card_id` совпадает
- карты игрока и карты оппонента никогда не смешиваются даже при одинаковом `card_id`

В статистику включаются только события, удовлетворяющие условиям:

- событие имеет тип `card_played`
- у события есть непустой `card_id`
- карта относится к стороне `self` или `opponent`

По умолчанию из этапа исключаются:

- hero power activations
- effect-only события без прямого `CardPlayedEvent`
- события без определимого `card_id`

## 5. Card Metadata Policy

Клиенту нужен минимальный карточный payload. На уровне спецификации он определяется так:

- `name` - локализованное имя карты
- `description` - очищенный и человекочитаемый текст карты
- `image` - ссылка или путь к изображению карты

### 5.1 Name

`name` берется из `carddb.CardInfo.Name`.

Fallback:

- если карта не найдена в базе, `name = card_id`

### 5.2 Description

`description` берется из `carddb.CardInfo.Text`, но обязательно проходит ту же очистку, что и текущий JSON formatter:

- удаление HTML tags
- удаление `[x]`
- нормализация переводов строк
- decode HTML entities

Fallback:

- если карта не найдена в базе или у нее нет текста, `description = ""`

### 5.3 Image

Текущая локальная `data/cards.json` не содержит готового image URL/path. Поэтому картинка не должна считаться прямым полем card database.

Для этой стадии вводится отдельная ответственность `CardAssetResolver`, которая получает как минимум:

- `card_id`
- `dbf_id` если доступен

И возвращает одно из состояний:

- `url`
- `file`
- `none`

Ключевое правило:

- live card stats projector не должен сам содержать hardcoded knowledge о внешнем CDN/provider
- стратегия резолва картинки должна быть конфигурируемой и заменяемой

Fallback:

- если изображение не удалось разрешить, `image.kind = "none"`

## 6. Projection Contract

Проекция должна быть маленькой, детерминированной и пригодной для постоянной перезаписи.

Рекомендуемый snapshot contract:

```json
{
  "schema_version": "live_card_stats/v1",
  "projection_version": 1,
  "recovery_status": "fresh",
  "match": {
    "key": "game:123",
    "status": "active",
    "started_at": "2026-04-27T12:00:00Z",
    "updated_at": "2026-04-27T12:03:15Z"
  },
  "players": {
    "self": {
      "total_plays": 5,
      "unique_cards": 4,
      "cards": [
        {
          "card_id": "CS2_029",
          "play_count": 2,
          "first_play_seq": 1,
          "last_play_seq": 4,
          "card": {
            "name": "Огненный шар",
            "description": "Наносит 6 ед. урона.",
            "image": {
              "kind": "url",
              "value": "https://..."
            }
          }
        }
      ]
    },
    "opponent": {
      "total_plays": 3,
      "unique_cards": 3,
      "cards": []
    }
  }
}
```

### 6.1 Match Metadata

`match.status` принимает значения:

- `idle`
- `active`
- `finalized`

Правила:

- `idle` - активного матча еще нет
- `active` - матч идет, статистика обновляется
- `finalized` - матч завершен, финальный snapshot еще доступен клиенту

### 6.2 Per-side Stats

Для каждой стороны хранятся:

- `total_plays` - общее число зафиксированных `CardPlayedEvent`
- `unique_cards` - число уникальных `card_id`
- `cards[]` - агрегированные карточные записи

Явное требование к структуре snapshot:

- `players.self.cards` и `players.opponent.cards` являются двумя разными массивами
- единый общий массив карт матча с признаком стороны не допускается

### 6.3 Card Entry

Минимальная агрегированная запись должна содержать:

- `card_id`
- `play_count`
- `first_play_seq`
- `last_play_seq`
- `card.name`
- `card.description`
- `card.image`

`first_play_seq` и `last_play_seq` нужны не столько клиенту, сколько для:

- детерминированной сортировки
- future delta updates
- корректной отладки live-потока

## 7. Lifecycle

### 7.1 GameStart

При `GameStart` проекция:

- сбрасывает статистику предыдущего матча
- открывает новый `match.status = active`
- инициализирует пустые buckets `self` и `opponent`

### 7.2 CardPlayedEvent

При каждом `CardPlayedEvent` проекция:

- определяет bucket (`self` или `opponent`)
- увеличивает `total_plays`
- создает или обновляет запись по `card_id`
- повышает `play_count`
- обновляет `last_play_seq`
- публикует новый snapshot

### 7.3 GameEnd

При `GameEnd` проекция:

- не очищает накопленные данные
- переключает `match.status = finalized`
- публикует финальный snapshot матча

### 7.4 Next GameStart

Следующий `GameStart` закрывает доступность предыдущей live-проекции и начинает новый цикл.

## 8. Delivery Model

Core projection logic должна быть transport-agnostic.

Вводится абстракция уровня sink:

- `LiveCardStatsSink.Publish(snapshot)`

Для первого этапа обязательным delivery surface считается:

- атомарно перезаписываемый JSON snapshot file, например `output/live_card_stats.json`

Почему именно так:

- проект уже CLI-first
- file snapshot проще всего подключить к внешнему клиенту
- это не привязывает архитектуру к веб-серверу
- file transport можно позже заменить или дополнить без изменения projector logic

Требования к snapshot file:

- запись только при meaningful change
- overwrite, а не append
- atomic replace через temp file + rename

## 9. Recovery And Restart Semantics

Для live client projection важна честность при рестарте monitor процесса.

Проблема:

- текущий `monitor state` хранит offsets reader-ов, но не хранит полную active card stats projection
- если процесс будет перезапущен посреди матча, клиентская статистика может стать partial или быть потеряна

На этом этапе нужно зафиксировать одно из двух допустимых поведений:

### Вариант A: Resume-capable projection

Система сохраняет:

- active projection snapshot
- match metadata
- source offsets, по которым snapshot был построен

И умеет после рестарта продолжить накопление без двойного счета.

### Вариант B: Honest partial reset

Если точное продолжение после рестарта не реализовано, система обязана:

- сбросить live projection для активного матча
- выставить `recovery_status = "partial"` или аналогичный явный статус
- не выдавать молча неполную статистику как полную

Ключевое правило:

- silent corruption live stats после рестарта недопустима

## 10. Architecture Integration

Новый этап должен добавить отдельный lightweight runtime component, например:

- `LiveCardStatsProjector`

Его ответственность:

- принимать live events
- поддерживать in-memory aggregation
- enrich-ить card payload
- публиковать snapshot в sink

Он не должен:

- заменять `SessionDetector`
- зависеть от финализации матча в canonical storage
- дублировать heavy replay formatter

Правильная роль компонента:

- secondary live projection для клиента

## 11. Non-Functional Requirements

- обновление проекции должно быть near-real-time относительно `CardPlayedEvent`
- проекция не должна требовать reread всего матча на каждое обновление
- memory footprint должен расти только с числом уникальных сыгранных карт текущего матча, а не с длиной всего event log
- отсутствие картинки или carddb lookup не должно ломать публикацию snapshot

## 12. Tests

Обязательный baseline:

1. `TestLiveCardStatsAggregatesBySideAndCardID`
2. `TestLiveCardStatsKeepsSelfAndOpponentInSeparateArrays`
3. `TestLiveCardStatsIncrementsPlayCountForRepeatedCard`
4. `TestLiveCardStatsUsesCardDBNameAndCleanDescription`
5. `TestLiveCardStatsFallsBackWhenCardMetadataMissing`
6. `TestLiveCardStatsIgnoresEventsWithoutPlayableCardID`
7. `TestLiveCardStatsResetsOnGameStart`
8. `TestLiveCardStatsFinalizesOnGameEndWithoutDataLoss`
9. `TestLiveCardStatsDoesNotLeakHiddenOpponentInformation`
10. `TestLiveCardStatsSnapshotFileAtomicReplace`
11. `TestLiveCardStatsRecoveryStatusAfterRestart`

## 13. Критерий завершения Task 13

Task `13` считается завершенным, когда:

1. определен продуктовый scope live card stats проекции
2. зафиксирован snapshot contract для клиента
3. зафиксированы правила aggregation по `self` и `opponent`
4. зафиксирована отдельная image resolution policy
5. определена честная recovery semantics
6. определен обязательный набор unit/integration/regression tests
7. для каждого пункта спецификации написаны и проходят соответствующие тесты
