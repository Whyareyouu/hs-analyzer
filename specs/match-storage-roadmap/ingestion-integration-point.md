# Ingestion Integration Point (Task 8.1)

## Цель

Определить, где и как интегрировать запись canonical match storage в текущий pipeline.

## 1. Выбранная точка интеграции

Рекомендуемая точка записи:

- после финализации матча (`match finalized`)
- на уровне post-game pipeline / finalize стадии analyzer workflow

Почему:

- матч уже сегментирован и имеет стабильные границы
- можно обеспечить транзакционную запись единицей `one finalized match`
- проще обеспечить idempotency

## 2. Контракт между `state.State` и storage writer

Минимальный контракт входа writer:

- `match identity` (game_id/fallback key)
- список canonical events в стабильном порядке
- turn segmentation metadata
- reconstructed/derived metadata (mode, players summary, completion status)

`state.State` не является source of truth, а источником для нормализованной canonical serialization.

## 3. Переходный период (dual output)

На переходном этапе поддерживается:

- запись в canonical storage (новый путь)
- существующий formatter output (`timeline_*`) как compatibility/export слой

Правило:

- canonical write failure не должен silently игнорироваться
- состояние dual path должно быть диагностируемым

## 4. Integration tests

1. `TestFinalizeMatchTriggersCanonicalWrite`
2. `TestFinalizeMatchIdempotentWrite`
3. `TestDualPathCanonicalAndCompatibilityOutput`
4. `TestCanonicalWriteFailureIsReported`

## 5. Критерий завершения 8.1

Шаг `8.1` считается завершенным, когда:

1. Выбрана точка интеграции записи canonical storage.
2. Зафиксирован контракт между `state.State` и storage writer.
3. Зафиксирована стратегия переходного dual-output периода.
4. Определены integration tests для выбранной точки интеграции.
