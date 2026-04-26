# Decision Point Definition (Task 7.1)

## Цель

Определить формальные правила, что считается decision point в матчевых логах.

## 1. Базовое определение

`DecisionPoint` - момент, когда у игрока есть осмысленный выбор действия, влияющий на последующее состояние и/или исход.

## 2. Типы decision points

- `start_turn_decision`
- `post_draw_decision`
- `discover_choice_decision`
- `attack_sequence_decision`
- `end_turn_decision`

Дополнительные типы могут быть добавлены versioned-расширением.

## 3. Правила выделения

Decision point выделяется, если:

1. Игрок имеет >1 допустимого action path.
2. Выбор приводит к различимым state/action/outcome веткам.
3. Момент может быть привязан к canonical turn/event позиции.

Не выделяется, если:

- событие строго автоматическое и без выбора игрока
- это лишь технический parser artifact

## 4. Отличие от обычного event node

- `Event` фиксирует факт, который произошел.
- `DecisionPoint` фиксирует пространство выбора до факта действия.

Связь:

- `DecisionPoint -> CandidateActions -> ChosenAction -> Outcomes`

## 5. Идентификация в логах

Минимальные сигналы:

- границы хода (`turn start/end`)
- `discover` и аналогичные выборы
- начало/порядок атак
- явные действия игрока по розыгрышу/таргету/пассу

## 6. Tests

1. `TestDetectStartTurnDecisionPoints`
2. `TestDetectPostDrawDecisionPoints`
3. `TestDetectDiscoverDecisionPoints`
4. `TestDetectAttackSequenceDecisionPoints`
5. `TestDetectEndTurnDecisionPoints`
6. `TestNoDecisionPointForForcedAutomaticEvents`

## 7. Критерий завершения 7.1

Шаг `7.1` считается завершенным, когда:

1. Определено формальное понятие decision point.
2. Зафиксированы типы decision points и правила выделения.
3. Зафиксировано отличие decision point от обычного event node.
4. Определены тесты на корректное выделение в типовых сценариях.
