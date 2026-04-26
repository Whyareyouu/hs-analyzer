# Decision Dataset Foundation (Task 7.3)

## Цель

Подготовить требования к данным для будущего сравнения выбранного хода с альтернативами (эвристики, ML, rule-based анализ).

## 1. Извлекаемые признаки

Базовые feature-группы:

- state features (ресурсы, board pressure, tempo proxies, hand profile)
- action features (тип действия, цель, стоимость, порядок)
- outcome features (краткосрочные и отложенные эффекты)
- match context features (mode, archetype hints, turn stage)

## 2. Offline dataset generation

Контракт:

- генерация датасета из canonical storage без зависимостей от текущего runtime formatter
- versioned schema для dataset rows
- поддержка incremental export по диапазону `match_id`/времени

## 3. Сравнение chosen action и alternatives

Для каждого decision point:

- фиксируется `chosen_action`
- фиксируются `candidate_actions`
- рассчитываются сравнимые outcome метрики на одинаковом горизонте (например immediate board delta / short horizon value proxy)

Если альтернативы неполные:

- обязательно ставится флаг `alternatives_incomplete`

## 4. Dataset validation tests

1. `TestDatasetRowContainsCoreDecisionFields`
2. `TestDatasetRowsReferenceCanonicalIDs`
3. `TestDatasetGenerationDeterminism`
4. `TestDatasetAlternativeComparisonFields`
5. `TestDatasetIncompleteAlternativesFlag`

## 5. Feature extraction regression tests

1. `TestFeatureExtractionStateFeaturesRegression`
2. `TestFeatureExtractionActionFeaturesRegression`
3. `TestFeatureExtractionOutcomeFeaturesRegression`
4. `TestFeatureExtractionNoHiddenLeakInPlayerViewMode`

## 6. Критерий завершения 7.3

Шаг `7.3` считается завершенным, когда:

1. Определены извлекаемые признаки под future decision analysis.
2. Определен контракт offline dataset generation.
3. Зафиксированы правила сравнения chosen action и alternatives.
4. Определены dataset validation и feature extraction regression tests.
