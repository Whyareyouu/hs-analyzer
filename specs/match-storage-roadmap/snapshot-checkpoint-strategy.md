# Snapshot and Checkpoint Strategy (Task 5.1)

## Цель

Выбрать стратегию snapshots/checkpoints, которая балансирует размер хранения и скорость реконструкции состояния.

## 1. Сравнение стратегий

## 1.1 Snapshot на каждом ходе

Плюсы:

- быстрая реконструкция позиции в начале/конце хода
- простая модель чтения

Минусы:

- значительный рост размера хранилища
- дублирование данных при малых изменениях состояния

## 1.2 Snapshot на каждом decision point

Плюсы:

- точная аналитическая привязка к выбору игрока
- удобство для explainability и dataset generation

Минусы:

- может быть слишком много snapshots в длинных/сложных матчах
- сложнее контролировать storage footprint

## 1.3 Sparse checkpoints (редкие контрольные точки)

Плюсы:

- минимальный размер хранения

Минусы:

- дорогое восстановление дальних позиций
- повышенная задержка materialization

## 1.4 Выбранный гибрид

Рекомендуемая стратегия:

- обязательный snapshot на границах хода (`turn_start`/`turn_end`)
- optional snapshot на важных `decision_point` (discover, attack-sequence pivot, end-turn decision)
- fallback sparse checkpoint каждые `N` событий для длинных ходов

Это дает:

- контролируемый размер
- предсказуемую latency реконструкции
- достаточный контекст для анализа решений

## 2. Минимальный state snapshot

Snapshot должен содержать достаточно данных для анализа решения:

- активный/неактивный игрок
- здоровье/броня героев
- ресурсы (`mana current/max`, overload-like статусы при наличии)
- публичный board state
- рука текущего игрока в рамках разрешенной visibility-модели
- graveyard/discard summary
- ключевые статусы сущностей (taunt, stealth, frozen, exhausted и т.п.)

Snapshot не должен включать:

- повтор полной истории событий
- неограниченные debug blobs без флага

## 3. Компрессия и дедупликация snapshot-данных

Правила:

- snapshots хранятся как структурированные данные с field-level дедупликацией неизменных сегментов
- большие повторяющиеся подструктуры (например card metadata) выносятся в словари/ref
- допускается компрессия payload на уровне storage, если не ломает random-access

## 4. Tests для полноты и восстановления

1. `TestSnapshotCompletenessForDecisionAnalysis`
   - snapshot содержит минимальный обязательный контекст.
2. `TestSnapshotDedupDoesNotLoseSemantics`
   - дедупликация не меняет восстановленное состояние.
3. `TestCheckpointRecoveryDeterminism`
   - одинаковый checkpoint + event stream всегда дает один и тот же state.
4. `TestSparseCheckpointLongTurnRecovery`
   - реконструкция длинного хода остается корректной при sparse checkpoint.
5. `TestVisibilityScopeInSnapshots`
   - hidden/public данные разделены согласно visibility policy.

## 5. Критерий завершения 5.1

Шаг `5.1` считается завершенным, когда:

1. Сравнены базовые snapshot/checkpoint стратегии.
2. Зафиксирована выбранная гибридная стратегия.
3. Определен минимальный snapshot contract для анализа решений.
4. Зафиксированы правила компрессии/дедупликации и тесты на корректность.
