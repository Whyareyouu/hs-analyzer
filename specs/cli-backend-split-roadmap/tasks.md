# Implementation Plan

## Правило приемки

- Для каждого пункта и подпункта implementation plan должны быть определены и реализованы соответствующие тесты.
- Корректно выполненные тесты являются обязательной основой приемки задачи.
- Задача не считается завершенной, если код реализован, но отсутствуют или не проходят тесты, подтверждающие ее поведение.
- Для каждого этапа необходимо заранее определить уровень проверки: unit, integration, contract, regression, recovery или performance test.

- [ ] 1. Сформировать архитектурное решение для split `hs-analyzer-cli` + `hs-analyzer-be`

  - Зафиксировать, что текущая локальная реализация становится `hs-analyzer-cli`
  - Зафиксировать, что backend становится владельцем canonical storage, projections, materialized views, history, analytics и future best-move
  - Зафиксировать, что CLI больше не владеет SQLite/graph migrations, rebuild projections и тяжелой аналитикой
  - Зафиксировать, что CLI сохраняет только durable delivery buffer, а не product source of truth
  - Подготовить architecture acceptance tests или static dependency tests для проверки направления зависимостей
  - _Requirements: 1, 2, 3, 6, 8_

  - [ ] 1.1 Определить ownership текущих компонентов

    - Разнести текущие компоненты по категориям: CLI-owned, backend-owned, shared contracts
    - Зафиксировать, что parser/live monitor/session detector остаются в CLI
    - Зафиксировать, что storage schema/materialization/projections уходят на backend
    - Подготовить tests/checks на forbidden imports между CLI и backend

  - [ ] 1.2 Зафиксировать целевую структуру папок

    - Подготовить target layout `apps/hs-analyzer-cli`, `services/hs-analyzer-be`, `pkg/contracts`
    - Описать поэтапный перенос текущего `cmd/hsanalyzer` и internal packages
    - Зафиксировать import rules для shared contracts
    - Подготовить static tests/checks на корректную структуру зависимостей

- [ ] 2. Спроектировать и реализовать `match_upload/v1` contract

  - Определить DTO для versioned match payload
  - Определить обязательные поля: `schema_version`, `match_id`, `completion_status`, `completion_reason`, `source`, `events`
  - Определить optional поля: `game_id`, timestamps, metadata, snapshots
  - Зафиксировать, что raw logs не являются primary upload contract
  - Подготовить contract tests на valid/invalid payloads
  - _Requirements: 4, 9_

  - [ ] 2.1 Определить idempotency contract

    - Зафиксировать `Idempotency-Key`
    - Определить порядок выбора key: `match_id`, reliable `game_id`, event stream hash
    - Зафиксировать behavior для duplicate same payload и duplicate different payload
    - Подготовить contract tests на duplicate/idempotency сценарии

  - [ ] 2.2 Определить API response/error contract

    - Зафиксировать `202 accepted`, `200 stored`, `409 duplicate_conflict`, `422 validation_failed`, `503 unavailable`
    - Зафиксировать JSON error/response format
    - Подготовить tests на response schema и retry classification

- [ ] 3. Реализовать CLI-side completed match package

  - Интегрировать сбор match package с текущим live session finalization
  - Сформировать payload из completed session events и reconstructed state
  - Гарантировать отправку package после завершения каждой игры
  - Сохранить независимость local live snapshot от upload path
  - Подготовить integration tests на `GameEnd -> package built`
  - _Requirements: 2, 4, 7, 9_

  - [ ] 3.1 Заменить runtime-only persistence boundary на delivery boundary

    - Переосмыслить текущий `MatchPersister` как uploader/delivery interface
    - Убрать предположение, что CLI пишет в canonical DB
    - Подготовить tests на successful upload handoff и skipped invalid session

  - [ ] 3.2 Зафиксировать fallback finalization uploads

    - Определить payload для `next_game_start`, `flush`, `stream_end`
    - Зафиксировать `complete` vs `incomplete`
    - Подготовить regression tests на incomplete session upload policy

- [ ] 4. Реализовать durable delivery buffer в CLI

  - Создать local spool для pending match uploads
  - Реализовать retry с deterministic test clock
  - Реализовать восстановление очереди после restart
  - Реализовать handling retryable/non-retryable errors
  - Подготовить recovery tests на server unavailable и CLI restart
  - _Requirements: 5, 9_

  - [ ] 4.1 Реализовать spool item lifecycle

    - Определить статусы `pending`, `sending`, `delivered`, `retry_wait`, `rejected`
    - Гарантировать write-before-send
    - Подготовить tests на state transitions и delivered cleanup

  - [ ] 4.2 Реализовать idempotent retry

    - Всегда отправлять `Idempotency-Key`
    - Повторять retryable errors
    - Не повторять validation/rejected errors бесконечно
    - Подготовить tests на retry classifier и idempotent resend

- [ ] 5. Создать backend ingestion service

  - Создать `services/hs-analyzer-be`
  - Реализовать `POST /api/v1/matches`
  - Подключить validation для `match_upload/v1`
  - Подключить idempotency behavior
  - Подготовить API integration tests
  - _Requirements: 3, 4, 6, 9_

  - [ ] 5.1 Реализовать backend validation layer

    - Проверять schema version, required fields, enum values и payload invariants
    - Возвращать стабильные API errors
    - Подготовить validation contract tests

  - [ ] 5.2 Реализовать backend idempotency layer

    - Хранить idempotency key и payload hash
    - Возвращать stored/accepted для повторов того же payload
    - Возвращать conflict для другого payload с тем же key
    - Подготовить idempotency integration tests

- [ ] 6. Перенести canonical storage ownership на backend

  - Добавить Postgres storage layer в backend
  - Перенести migrations ownership в backend
  - Сохранять canonical match/events из upload payload
  - Исключить CLI ownership SQL/graph schema migrations
  - Подготовить storage integration tests
  - _Requirements: 3, 6, 9_

  - [ ] 6.1 Реализовать backend match writer

    - Записывать match metadata и events в стабильном порядке
    - Поддерживать idempotent write
    - Подготовить readback tests

  - [ ] 6.2 Зафиксировать server-side migration strategy

    - Хранить migrations рядом с backend storage layer
    - Зафиксировать compatibility policy для schema changes
    - Подготовить migration tests

- [ ] 7. Реализовать server-side projections и materialized views

  - Построить materialized match summary из canonical storage
  - Построить graph projection как server-owned projection layer
  - Реализовать deterministic rebuild
  - Подготовить projection integrity tests
  - _Requirements: 3, 6, 9_

  - [ ] 7.1 Реализовать materialized match views

    - Подготовить view для match history
    - Подготовить view для match detail/materialized timeline
    - Подготовить tests на equivalence с canonical data

  - [ ] 7.2 Реализовать graph projection foundation

    - Построить graph nodes/edges из canonical IDs
    - Исключить graph-only facts, которые нельзя восстановить
    - Подготовить rebuild/determinism tests

- [ ] 8. Сохранить local Electron live boundary

  - Оставить live snapshot delivery независимым от backend upload
  - Зафиксировать, что backend downtime не ломает local UI
  - Подготовить tests на live snapshot при server unavailable
  - _Requirements: 2, 7, 9_

  - [ ] 8.1 Подготовить future WebSocket delivery для local UI

    - Зафиксировать, что WebSocket является local UI transport, а не canonical storage
    - Подготовить tests на initial snapshot и update push
    - Не блокировать match upload rollout этим transport

- [ ] 9. Подготовить migration strategy от текущего roadmap состояния

  - Обновить смысл `match-storage-roadmap`: реализация storage находится на backend стороне
  - Зафиксировать, какие текущие документы остаются актуальными как server storage design
  - Зафиксировать, что текущий CLI transition сначала добавляет uploader/spool, а не локальную DB
  - Подготовить regression tests на отсутствие поломки existing analyze/monitor flows
  - _Requirements: 1, 6, 8, 9_

- [ ] 10. Подготовить end-to-end acceptance flow

  - Запустить live monitor
  - Завершить матч
  - Сформировать upload payload
  - Доставить на backend или сохранить в spool при недоступности
  - Backend валидирует и сохраняет матч
  - Backend строит projection/materialized view
  - Подготовить full integration или staged integration tests
  - _Requirements: 1-9_

