# Primary Storage Decision (Task 2.2)

## Контекст

Сравнение выполнено по критериям из `storage-selection-criteria.md`:

- локальный запуск и эксплуатация
- транзакционная целостность записи матча
- скорость чтения `match` и `turn`
- миграции и versioning
- дебаг/backup/переносимость
- пригодность для аналитических выборок и rebuild проекций

Шкала: `0..5`, итоговый score - взвешенная сумма.

## Кандидаты

1. `SQLite` (embedded SQL, WAL)
2. `PostgreSQL` (внешний сервер)
3. `Embedded KV` (Badger/Bbolt-class)
4. `Append-only file storage` (кастомный формат)

## Матрица оценки

| Критерий | Вес | SQLite | PostgreSQL | Embedded KV | Append-only file |
|---|---:|---:|---:|---:|---:|
| Локальный запуск и эксплуатация | 20% | 5 | 2 | 4 | 5 |
| Транзакционная целостность и идемпотентность | 20% | 5 | 5 | 3 | 2 |
| Скорость чтения `match` и `turn` | 20% | 4 | 4 | 3 | 2 |
| Миграции и versioning | 15% | 4 | 5 | 2 | 1 |
| Дебаг/backup/переносимость | 15% | 5 | 3 | 2 | 2 |
| Аналитические выборки и projection rebuild | 10% | 4 | 5 | 2 | 1 |
| **Итоговый score** | 100% | **4.6** | **3.9** | **2.9** | **2.5** |

## Решение

### Recommended primary storage: `SQLite`

Причины:

- Наилучший баланс локального DX и production-grade транзакционности.
- Простые и проверяемые миграции схемы.
- Нативный SQL для точечных выборок (`match`, `turn`, диапазон `events`).
- Низкая операционная стоимость: один файл, удобный backup/перенос.
- Хорошая основа для deterministic rebuild вторичных проекций.

### Роль остальных кандидатов

- `PostgreSQL`: допустим как future optional backend для server deployment, но не как базовый primary storage для текущего локального workflow.
- `Embedded KV`: возможен как low-level storage engine, но потребует существенно больше custom-логики для индексов, выборок и миграций.
- `Append-only file`: полезен как debug/export или WAL-подобный артефакт, но слаб как самостоятельный primary source of truth.

## Подтверждение выбора (reproducible tests/benchmarks)

Для закрытия implementation-этапа по runtime-коду (не только spec) обязательны:

1. Smoke tests:
   - запись одного матча
   - идемпотентная повторная запись
   - чтение `match` по `match_id`
   - чтение `turn` без materialization всего матча
2. Integration tests:
   - `logs -> analyzer -> storage writer -> storage reader -> materializer`
   - rebuild projection из SQLite без потери целостности
3. Benchmarks:
   - запись `small/medium/long` матча
   - чтение `match`, `turn`, `event range`
   - измерение storage footprint на 1/100/1000 матчей

Порог приемки:

- отсутствие функциональных regressions на smoke/integration
- стабильная работа migration tests
- p95 latencies и storage footprint должны быть не хуже зафиксированных целевых baseline (фиксируются в задаче 11)
