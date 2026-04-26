# Transactional Match Writer (Task 8.2)

## Цель

Определить транзакционную и идемпотентную запись одного матча в primary storage.

## 1. Единица транзакции

Одна транзакция = один матч:

- `match`
- `turns`
- `events`
- `snapshots` (если есть)
- `entities`
- `decision_points` (если есть)

При ошибке любой части запись откатывается целиком.

## 2. Идемпотентность

Ключ идемпотентности:

- `match_id` + `schema_version`

Повторная запись того же матча:

- не должна создавать дубликаты
- должна возвращать статус `already_persisted` или эквивалент

## 3. Partially parsed games

Поддерживается запись incomplete матча:

- `completion_status = incomplete`
- сохранение доступных фактов без искажения canonical истории
- флаг качества/полноты данных для downstream аналитики

## 4. Ошибки и retry

- transient storage errors -> retry policy
- non-recoverable contract errors -> явная ошибка без partial commit
- retry не должен нарушать идемпотентность

## 5. Tests

1. `TestWriterTransactionalCommitAllOrNothing`
2. `TestWriterIdempotentMatchSave`
3. `TestWriterHandlesIncompleteMatch`
4. `TestWriterRetryOnTransientFailure`
5. `TestWriterNoDuplicateOnRetry`

## 6. Критерий завершения 8.2

Шаг `8.2` считается завершенным, когда:

1. Зафиксирована транзакционная единица записи матча.
2. Зафиксированы правила идемпотентности.
3. Определена стратегия для partially parsed/incomplete матчей.
4. Определены тесты на транзакционность, retry и отсутствие дублей.
