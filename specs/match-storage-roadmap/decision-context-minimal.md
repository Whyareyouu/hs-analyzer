# Minimal Decision Context (Task 7.2)

## Цель

Определить минимальный контекст, который должен сопровождать decision point для честного анализа качества хода.

## 1. Обязательные признаки контекста

- `player_ref` (кто принимает решение)
- `turn_number`
- `phase` decision point
- ресурсы игрока (`mana`, доступные модификаторы)
- рука игрока (в рамках разрешенной visibility)
- публичный board state
- graveyard/discard summary
- известная информация об оппоненте (только доступная на момент решения)

## 2. Hidden/uncertain данные

Запрещено использовать как ground truth для честной оценки:

- скрытые карты оппонента
- недоступные игроку внутренние состояния, не раскрытые в матче

Допускается хранить отдельно:

- `hidden_context` для offline-research режима
- но по умолчанию evaluation использует `player_view`

## 3. Candidate actions и chosen action

Для каждого decision point:

- `candidate_actions[]`
- `chosen_action`
- `action_extraction_confidence`

Если список кандидатов неполный:

- это явно помечается флагом `candidate_set_incomplete`

## 4. Связь с snapshot и outcome

Decision point должен ссылаться на:

- `snapshot_before_ref`
- `chosen_action_ref`
- `outcome_window_ref` (набор событий/состояний после выбора)

## 5. Tests

1. `TestDecisionContextContainsRequiredFields`
2. `TestDecisionContextNoHiddenLeakageInPlayerView`
3. `TestDecisionContextCandidateActionsPresence`
4. `TestDecisionContextSnapshotOutcomeLinkIntegrity`
5. `TestDecisionContextIncompleteCandidateSetFlag`

## 6. Критерий завершения 7.2

Шаг `7.2` считается завершенным, когда:

1. Определен обязательный набор decision context признаков.
2. Зафиксированы правила разделения hidden/public данных.
3. Зафиксирован контракт для candidate/chosen actions.
4. Зафиксированы связи decision point со snapshot и outcome.
5. Определены тесты на полноту и отсутствие утечки скрытой информации.
