# Requirements Document

## Introduction

`cli-backend-split-roadmap` определяет переход от монолитного локального приложения к разделенной архитектуре:

- `hs-analyzer-cli` - локальный ingestion agent для парсинга логов, live-monitoring, сборки завершенного match package и доставки данных на сервер.
- `hs-analyzer-be` - backend-сервис, который владеет приемом матчей, валидацией, canonical storage, server-side projections, materialized views, историей матчей, аналитикой и future best-move.

Цель разделения - оставить на клиенте только то, что требует доступа к локальным логам Hearthstone и live UI, а всю тяжелую storage/analytics логику перенести на сервер.

## Requirements

### Requirement 1: Разделение ответственности CLI и backend

**User Story:** Как разработчик, я хочу четко разделить локальный агент и серверную часть, чтобы desktop-приложение не владело storage schema, migrations, graph rebuild и тяжелой аналитикой.

#### Acceptance Criteria

1. WHEN проект переходит на split-архитектуру THEN текущая локальная runtime-часть SHALL быть описана как `hs-analyzer-cli`
2. WHEN матч завершен THEN `hs-analyzer-cli` SHALL собрать versioned match package и передать его на backend
3. WHEN требуется canonical storage THEN source of truth SHALL находиться на стороне `hs-analyzer-be`
4. WHEN требуется graph projection THEN она SHALL строиться на стороне backend из canonical storage
5. WHEN требуется history/materialized view/analytics/best-move THEN эти сценарии SHALL обслуживаться backend API
6. WHEN задача считается реализованной THEN соответствующие tests SHALL быть написаны и проходить

### Requirement 2: Ответственность `hs-analyzer-cli`

**User Story:** Как пользователь локального приложения, я хочу, чтобы CLI продолжал читать мои Hearthstone logs и показывать live-данные, даже если сервер временно недоступен.

#### Acceptance Criteria

1. WHEN CLI запускается THEN он SHALL искать и читать локальные Hearthstone logs
2. WHEN CLI получает live события THEN он SHALL поддерживать parser, live monitor и session detector
3. WHEN CLI получает `CardPlayedEvent` THEN он SHALL обновлять live snapshot для Electron/local UI
4. WHEN активный матч завершается THEN CLI SHALL собрать completed match package
5. WHEN backend недоступен THEN CLI SHALL сохранить pending upload в durable delivery buffer
6. WHEN backend снова доступен THEN CLI SHALL выполнить retry pending uploads
7. WHEN CLI перезапускается THEN pending uploads SHALL не теряться

### Requirement 3: Ответственность `hs-analyzer-be`

**User Story:** Как разработчик backend, я хочу, чтобы сервер был единственным владельцем долговременного хранения и аналитики, чтобы развивать storage независимо от desktop-клиента.

#### Acceptance Criteria

1. WHEN backend получает match payload THEN он SHALL валидировать schema/version/content
2. WHEN payload валиден THEN backend SHALL записать матч в обычную БД
3. WHEN canonical storage записан THEN backend SHALL построить или запланировать graph/materialized projections
4. WHEN пользователь запрашивает историю матчей THEN backend SHALL читать ее из server-side storage
5. WHEN пользователь запрашивает match viewer/materialized view THEN backend SHALL строить представление из canonical storage
6. WHEN требуется аналитика или best-move THEN backend SHALL использовать server-side canonical/projection data

### Requirement 4: Versioned match upload contract

**User Story:** Как разработчик API, я хочу стабильный контракт загрузки завершенного матча, чтобы CLI и backend могли независимо развиваться.

#### Acceptance Criteria

1. WHEN CLI отправляет матч THEN payload SHALL иметь `schema_version`
2. WHEN CLI отправляет матч THEN payload SHALL содержать `match_id`, timestamps, completion metadata, events и minimal metadata
3. WHEN raw logs доступны локально THEN они SHALL NOT быть основным upload contract
4. WHEN server получает неизвестную `schema_version` THEN он SHALL вернуть validation error
5. WHEN payload повторно отправлен с тем же idempotency key THEN server SHALL не создавать дубль
6. WHEN payload отличается при том же idempotency key THEN server SHALL вернуть conflict

### Requirement 5: Durable delivery buffer

**User Story:** Как пользователь, я хочу не потерять завершенный матч, если сеть или сервер временно недоступны.

#### Acceptance Criteria

1. WHEN upload failed по transient причине THEN CLI SHALL сохранить match package в local spool
2. WHEN CLI получает retryable error THEN он SHALL выполнить retry с backoff
3. WHEN CLI получает non-retryable validation error THEN он SHALL пометить package как rejected и сохранить диагностический статус
4. WHEN CLI успешно доставил package THEN он SHALL удалить или архивировать spool item
5. WHEN CLI перезапускается THEN он SHALL восстановить pending queue из spool
6. WHEN queue содержит один и тот же матч несколько раз THEN idempotency SHALL защитить backend от duplicate save

### Requirement 6: Server-side storage ownership

**User Story:** Как разработчик, я хочу убрать из клиента ownership SQLite/graph storage, чтобы storage evolution не требовал обновления desktop-приложения.

#### Acceptance Criteria

1. WHEN проект реализует обычную БД THEN migrations SHALL жить в backend-сервисе
2. WHEN проект реализует graph projection THEN rebuild logic SHALL жить в backend-сервисе
3. WHEN проект реализует materialized views THEN они SHALL строиться на backend стороне
4. WHEN CLI формирует match package THEN он SHALL не зависеть от server DB schema
5. WHEN backend меняет storage schema THEN CLI SHALL оставаться совместимым при неизменном upload contract

### Requirement 7: Local UI и Electron boundary

**User Story:** Как разработчик UI, я хочу получать live данные от локального агента без ожидания server persistence.

#### Acceptance Criteria

1. WHEN Electron UI подключен локально THEN он SHALL получать live snapshot из CLI
2. WHEN backend недоступен THEN Electron live UI SHALL продолжать работать для текущего матча
3. WHEN матч завершен THEN CLI SHALL отправить match package на backend независимо от live UI
4. WHEN нужен remote real-time UI THEN это SHALL быть отдельным server-side или relay feature, не обязательным для v1 split

### Requirement 8: Repo layout и module ownership

**User Story:** Как разработчик, я хочу понятную структуру папок, чтобы не смешивать CLI runtime, backend API и shared contracts.

#### Acceptance Criteria

1. WHEN проект реорганизуется THEN локальное приложение SHALL находиться в `apps/hs-analyzer-cli`
2. WHEN backend добавляется THEN сервер SHALL находиться в `apps/hs-analyzer-be`
3. WHEN CLI и backend разделяют DTO THEN общие контракты SHALL находиться в `pkg/contracts`
4. WHEN package относится только к CLI THEN он SHALL не импортироваться backend-ом без явной причины
5. WHEN package относится только к backend storage/analytics THEN он SHALL не импортироваться CLI

### Requirement 9: Testing as acceptance

**User Story:** Как разработчик, я хочу принимать каждый пункт только по тестам, чтобы split не сломал live-monitoring и delivery.

#### Acceptance Criteria

1. WHEN task реализован THEN для него SHALL существовать unit, integration, contract, regression, recovery или performance test
2. WHEN tests отсутствуют THEN task SHALL считаться незавершенным
3. WHEN tests падают THEN task SHALL считаться незавершенным
4. WHEN меняется upload contract THEN contract tests SHALL быть обновлены до implementation
5. WHEN меняется delivery buffer THEN recovery/retry tests SHALL покрывать crash и server unavailable scenarios
