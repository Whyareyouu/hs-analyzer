# Phased Implementation Plan (Task 12)

## Цель

Зафиксировать поэтапную реализацию подсистемы с обязательными test gates для каждого этапа.

## Этап 1: Primary storage и canonical schema

Состав:

- утвердить и реализовать canonical schema
- реализовать базовые миграции
- реализовать минимальный storage reader/writer contracts

Test gate:

- schema contract unit tests
- migration regression tests

## Этап 2: Match-level compact writer (без graph projection)

Состав:

- транзакционный writer матча
- идемпотентность записи
- поддержка incomplete matches

Test gate:

- writer transactional/idempotency tests
- integration test finalize->persist

## Этап 3: Materializer и compatibility export

Состав:

- materializer/renderer контракты
- `text`, `human`, `replay` materialization
- compatibility режим с legacy timeline

Test gate:

- materializer contract tests
- compatibility regression tests

## Этап 4: Snapshots и точечное восстановление хода

Состав:

- snapshot/checkpoint стратегия
- reconstruction engine
- point-access для turn/event

Test gate:

- reconstruction invariant tests
- point-access integration tests

## Этап 5: Decision points и candidate actions

Состав:

- decision point extraction
- decision context
- candidate/chosen action linkage

Test gate:

- decision extraction regression tests
- hidden data leakage tests

## Этап 6: Graph projection для аналитики и explainability

Состав:

- minimal graph schema
- incremental/full rebuild
- explainability query support

Test gate:

- projection integrity tests
- query regression tests

## Этап 7: Foundation для анализа лучшего хода

Состав:

- dataset generation pipeline
- feature extraction foundation
- chosen vs alternatives comparison scaffolding

Test gate:

- dataset validation tests
- feature regression tests

## Общие правила приемки этапов

Этап не принимается, если:

- не закрыт обязательный test gate этапа
- отсутствуют тесты для новых контрактов/схем
- есть критические регрессии без согласованного исключения

## Критерий завершения 12

Шаг `12` считается завершенным, когда:

1. Зафиксированы этапы `1..7`.
2. Для каждого этапа определен обязательный test gate.
3. Зафиксировано правило, что этап не принимается без успешных тестов.
